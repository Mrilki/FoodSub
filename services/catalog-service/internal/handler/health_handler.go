package handler

import (
	"net/http"

	"github.com/Mrilki/catalog-service/pkg/logger"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	log *logger.Logger
}

func NewHealthHandler(log *logger.Logger) *HealthHandler {
	return &HealthHandler{log: log}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "catalog-service",
	})
}

func (h *HealthHandler) ReadyCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ready",
		"service": "catalog-service",
	})
}
