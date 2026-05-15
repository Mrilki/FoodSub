package handler

import (
	"net/http"

	"github.com/Mrilki/catalog-service/internal/service"

	"github.com/Mrilki/catalog-service/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ArchiveHandler struct {
	service service.ArchiveService
	log     *logger.Logger
}

func NewArchiveHandler(service service.ArchiveService, log *logger.Logger) *ArchiveHandler {
	return &ArchiveHandler{
		service: service,
		log:     log,
	}
}
func (h *ArchiveHandler) TriggerArchive(c *gin.Context) {
	ctx := c.Request.Context()

	olderThanDays := 90

	if err := h.service.ArchiveOldMenus(ctx, olderThanDays); err != nil {
		h.log.Error("Failed to trigger archive", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to trigger archive"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Archive process completed",
		"older_than_days": olderThanDays,
	})
}

func (h *ArchiveHandler) ListArchives(c *gin.Context) {
	ctx := c.Request.Context()

	archives, err := h.service.ListArchives(ctx)
	if err != nil {
		h.log.Error("Failed to list archives", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list archives"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": archives,
	})
}

func (h *ArchiveHandler) GetArchive(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename required"})
		return
	}

	ctx := c.Request.Context()
	data, err := h.service.GetArchive(ctx, filename)
	if err != nil {
		h.log.Error("Failed to get archive", zap.String("filename", filename), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Archive not found"})
		return
	}

	c.Data(http.StatusOK, "application/json", data)
}
