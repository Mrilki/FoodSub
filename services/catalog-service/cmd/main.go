package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Mrilki/catalog-service/internal/config"
	"github.com/Mrilki/catalog-service/internal/handler"
	"github.com/Mrilki/catalog-service/internal/kafka"
	"github.com/Mrilki/catalog-service/internal/middleware"
	"github.com/Mrilki/catalog-service/internal/repository"
	"github.com/Mrilki/catalog-service/internal/service"
	"github.com/Mrilki/catalog-service/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Config loaded successfully")

	loger, err := logger.New(cfg.Logging.Level)
	if err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer loger.Sync()
	fmt.Println("Logger initialized successfully")

	go func() {
		metricsRouter := gin.Default()
		metricsRouter.GET("/metrics", gin.WrapH(promhttp.Handler()))
		loger.Info("Starting Prometheus metrics server", zap.String("port", cfg.Observability.PrometheusPort))
		if err := metricsRouter.Run(":" + cfg.Observability.PrometheusPort); err != nil {
			loger.Error("Failed to start metrics server", zap.Error(err))
		}
	}()

	if cfg.Observability.JaegerEndpoint != "" {
		tracerProvider, err := middleware.InitTracer(
			cfg.Observability.ServiceName,
			cfg.Observability.JaegerEndpoint,
			loger,
		)
		if err != nil {
			loger.Error("Failed to init Jaeger tracer", zap.Error(err))
		} else {
			defer tracerProvider.Shutdown(context.Background())
			loger.Info("Jaeger tracing enabled")
		}
	}

	mongoClient, err := config.ConnectMongoDB(cfg.MongoDB.URI, cfg.MongoDB.Database)
	if err != nil {
		loger.Error("Failed to connect to MongoDB", zap.Error(err))
		os.Exit(1)
	}
	defer config.CloseMongoDB(mongoClient)

	redisClient, err := config.ConnectRedis(cfg.Redis.Addr, cfg.Redis.Password)
	if err != nil {
		loger.Error("Failed to connect to Redis", zap.Error(err))
		os.Exit(1)
	}
	defer config.CloseRedis(redisClient)

	var minioRepo repository.MinIORepository
	if cfg.MinIO.Endpoint != "" {
		minioRepo, err = repository.NewMinIORepository(
			cfg.MinIO.Endpoint,
			cfg.MinIO.AccessKey,
			cfg.MinIO.SecretKey,
			cfg.MinIO.Bucket,
			loger,
		)
		if err != nil {
			loger.Error("Failed to init MinIO repository", zap.Error(err))
		} else {
			loger.Info("MinIO repository initialized")
		}
	}
	menuRepo := repository.NewMongoMenuRepository(mongoClient, cfg.MongoDB.Database, loger)
	redisRepo := repository.NewRedisRepository(redisClient, cfg.Redis.CacheTTL, loger)

	var archiveService service.ArchiveService
	if minioRepo != nil {
		archiveService = service.NewArchiveService(menuRepo, minioRepo, loger)
		loger.Info("Archive service initialized")
	}
	menuService := service.NewMenuService(menuRepo, redisRepo, loger)
	loger.Info("Menu service initialized")

	var kafkaProducer kafka.KafkaProducer
	if cfg.Kafka.Brokers != "" {
		kafkaProducer, err = kafka.NewKafkaProducer(
			cfg.Kafka.Brokers,
			"menu.events",
			loger,
		)
		if err != nil {
			loger.Error("Failed to create Kafka producer", zap.Error(err))
		} else {
			defer kafkaProducer.Close()
			loger.Info("Kafka producer initialized")
		}
	}

	if cfg.Kafka.Brokers != "" && len(cfg.Kafka.ConsumerTopics) > 0 {
		kafkaConsumer, err := kafka.NewKafkaConsumer(
			cfg.Kafka.Brokers,
			cfg.Kafka.ConsumerGroup,
			cfg.Kafka.ConsumerTopics,
			loger,
		)
		if err != nil {
			loger.Error("Failed to create Kafka consumer", zap.Error(err))
		} else {
			defer kafkaConsumer.Stop()

			kafkaConsumer.RegisterHandler("order.scheduled", menuService.ProcessOrderScheduled)
			kafkaConsumer.RegisterHandler("order.delivered", menuService.ProcessOrderDelivered)

			ctx := context.Background()
			if err := kafkaConsumer.Start(ctx); err != nil {
				loger.Error("Failed to start Kafka consumer", zap.Error(err))
			} else {
				loger.Info("Kafka consumer started", zap.Strings("topics", cfg.Kafka.ConsumerTopics))
			}
		}
	}

	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()

	r.Use(middleware.LoggerMiddleware(loger))
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RateLimiterMiddleware(
		redisClient,
		cfg.RateLimit.Requests,
		time.Duration(cfg.RateLimit.Window)*time.Second,
		loger,
	))
	r.Use(middleware.MetricsMiddleware())

	if cfg.Observability.JaegerEndpoint != "" {
		r.Use(middleware.TracingMiddleware(cfg.Observability.ServiceName))
	}

	handler.RegisterRoutes(
		r,
		menuService,
		kafkaProducer,
		redisClient,
		cfg.RateLimit,
		cfg.JWT,
		archiveService,
		loger,
	)

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		loger.Info("Starting HTTP server", zap.String("port", cfg.Server.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			loger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	loger.Info("Shutting down Catalog Service...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		loger.Error("Server forced to shutdown", zap.Error(err))
	}

	loger.Info("Server exited gracefully")
}
