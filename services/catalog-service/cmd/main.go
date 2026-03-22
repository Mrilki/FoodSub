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
	"github.com/Mrilki/catalog-service/internal/middleware"
	"github.com/Mrilki/catalog-service/internal/repository"
	"github.com/Mrilki/catalog-service/internal/service"
	"github.com/Mrilki/catalog-service/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 1. Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Config loaded successfully")

	// 2. Инициализируем логгер
	loger, err := logger.New(cfg.Logging.Level)
	if err != nil {
		fmt.Printf("❌ Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer loger.Sync()
	fmt.Println("✅ Logger initialized successfully")

	// 3. Подключаемся к MongoDB
	mongoClient, err := config.ConnectMongoDB(cfg.MongoDB.URI, cfg.MongoDB.Database)
	if err != nil {
		loger.Error("Failed to connect to MongoDB", zap.Error(err))
		os.Exit(1)
	}
	defer config.CloseMongoDB(mongoClient)

	// 4. Создаём репозиторий
	menuRepo := repository.NewMongoMenuRepository(mongoClient, cfg.MongoDB.Database, loger)
	loger.Info("MongoDB repository initialized",
		zap.String("database", cfg.MongoDB.Database))

	// 5. ✅ Создаём SERVICE слой
	menuService := service.NewMenuService(menuRepo, loger)
	loger.Info("Menu service initialized")

	// 6. Создаём Gin роутер
	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()

	// 7. ✅ Подключаем middleware из отдельного пакета
	r.Use(middleware.LoggerMiddleware(loger))
	r.Use(middleware.CORSMiddleware())

	// 8. Регистрируем роуты
	handler.RegisterRoutes(r, menuService, loger)

	// 9. Запускаем HTTP сервер
	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	go func() {
		loger.Info("Starting HTTP server", zap.String("port", cfg.Server.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			loger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	fmt.Printf("🚀 Server running on http://localhost:%s\n", cfg.Server.Port)
	fmt.Println("📝 Test endpoints:")
	fmt.Println("   GET  http://localhost:" + cfg.Server.Port + "/health")
	fmt.Println("   GET  http://localhost:" + cfg.Server.Port + "/api/v1/menu")
	fmt.Println("   POST http://localhost:" + cfg.Server.Port + "/api/v1/admin/menu")

	// 10. Обработка сигнала завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	loger.Info("Shutting down Catalog Service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		loger.Error("Server forced to shutdown", zap.Error(err))
	}

	loger.Info("Server exited gracefully")
}
