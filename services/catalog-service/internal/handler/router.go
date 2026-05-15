package handler

import (
	"time"

	"github.com/Mrilki/catalog-service/internal/config"
	"github.com/Mrilki/catalog-service/internal/kafka"
	"github.com/Mrilki/catalog-service/internal/middleware"
	"github.com/Mrilki/catalog-service/internal/service"
	"github.com/Mrilki/catalog-service/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func RegisterRoutes(
	r *gin.Engine,
	menuService service.MenuService,
	producer kafka.KafkaProducer,
	redisClient *redis.Client,
	rateLimit config.RateLimitConfig,
	jwtConfig config.JWTConfig,
	archiveService service.ArchiveService,
	log *logger.Logger,
) {
	menuHandler := NewMenuHandler(menuService, producer, log)
	healthHandler := NewHealthHandler(log)
	archiveHandler := NewArchiveHandler(archiveService, log)

	r.GET("/health", healthHandler.HealthCheck)
	r.GET("/ready", healthHandler.ReadyCheck)

	r.GET("/api/v1/menu", menuHandler.GetAllMenus)
	r.GET("/api/v1/menu/:id", menuHandler.GetMenuByID)
	r.GET("/api/v1/menu/search", menuHandler.SearchMenus)

	admin := r.Group("/api/v1/admin")

	admin.Use(middleware.JWTMiddleware(
		jwtConfig.PublicKeyURL,
		jwtConfig.CacheTTL,
		log,
	))

	admin.Use(middleware.AdminRateLimiterMiddleware(
		redisClient,
		rateLimit.AdminRequests,
		time.Duration(rateLimit.AdminWindow)*time.Second,
		log,
	))

	admin.Use(middleware.RequireRole("ADMIN"))

	{
		admin.POST("/menu", menuHandler.CreateMenu)
		admin.PUT("/menu/:id", menuHandler.UpdateMenu)
		admin.DELETE("/menu/:id", menuHandler.DeleteMenu)

		admin.POST("/archive/trigger", archiveHandler.TriggerArchive)
		admin.GET("/archive", archiveHandler.ListArchives)
		admin.GET("/archive/:filename", archiveHandler.GetArchive)
	}
}
