package handler

import (
	"github.com/Mrilki/catalog-service/internal/service"
	"github.com/Mrilki/catalog-service/pkg/logger"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	r *gin.Engine,
	menuService service.MenuService,
	log *logger.Logger,
) {

	menuHandler := NewMenuHandler(menuService, log)
	healthHandler := NewHealthHandler(log)

	// Публичные endpoints
	r.GET("/health", healthHandler.HealthCheck)
	r.GET("/ready", healthHandler.ReadyCheck)

	r.GET("/api/v1/menu", menuHandler.GetAllMenus)
	r.GET("/api/v1/menu/:id", menuHandler.GetMenuByID)
	r.GET("/api/v1/menu/search", menuHandler.SearchMenus)

	//  Админские endpoints
	admin := r.Group("/api/v1/admin")
	{
		admin.POST("/menu", menuHandler.CreateMenu)
		admin.PUT("/menu/:id", menuHandler.UpdateMenu)
		admin.DELETE("/menu/:id", menuHandler.DeleteMenu)
	}
}
