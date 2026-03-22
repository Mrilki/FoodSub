package handler

import (
	"net/http"

	"github.com/Mrilki/catalog-service/pkg/logger"
	"github.com/gin-gonic/gin"
)

// HealthHandler — проверка здоровья сервиса
type HealthHandler struct {
	log *logger.Logger
}

// NewHealthHandler — конструктор
func NewHealthHandler(log *logger.Logger) *HealthHandler {
	return &HealthHandler{log: log}
}

// HealthCheck — GET /health
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "catalog-service",
	})
}

// ReadyCheck — GET /ready
func (h *HealthHandler) ReadyCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ready",
		"service": "catalog-service",
	})
}
