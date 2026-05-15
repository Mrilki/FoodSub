package handler

import (
	"net/http"
	"strings"

	"github.com/Mrilki/catalog-service/internal/kafka"
	"github.com/Mrilki/catalog-service/internal/model"
	"github.com/Mrilki/catalog-service/internal/service"
	"github.com/Mrilki/catalog-service/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type MenuHandler struct {
	service  service.MenuService
	producer kafka.KafkaProducer
	log      *logger.Logger
}

func NewMenuHandler(
	service service.MenuService,
	producer kafka.KafkaProducer,
	log *logger.Logger,
) *MenuHandler {
	return &MenuHandler{
		service:  service,
		producer: producer,
		log:      log,
	}
}

func (h *MenuHandler) GetAllMenus(c *gin.Context) {
	ctx := c.Request.Context()

	menus, err := h.service.GetAll(ctx)
	if err != nil {
		h.log.Error("Failed to get menus", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get menus"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": menus})
}

func (h *MenuHandler) GetMenuByID(c *gin.Context) {
	id := c.Param("id")

	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid menu ID"})
		return
	}

	ctx := c.Request.Context()
	menu, err := h.service.GetByID(ctx, id)
	if err != nil {
		h.log.Warn("Menu not found", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Menu not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": menu})
}

func (h *MenuHandler) CreateMenu(c *gin.Context) {
	var menu model.MenuItem

	if err := c.ShouldBindJSON(&menu); err != nil {
		h.log.Warn("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx := c.Request.Context()
	menu.BeforeCreate()

	if err := h.service.Create(ctx, &menu); err != nil {
		if _, ok := err.(*service.ValidationError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		h.log.Error("Failed to create menu", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create menu"})
		return
	}

	if h.producer != nil {
		if err := h.producer.SendMenuCreated(
			menu.ID.Hex(),
			menu.Name,
			menu.Category,
		); err != nil {
			h.log.Warn("Failed to send Kafka event", zap.Error(err))
			// Не прерываем запрос, событие опционально
		}
	}

	h.log.Info("Menu created", zap.String("id", menu.ID.Hex()))
	c.JSON(http.StatusCreated, gin.H{"data": menu})
}

func (h *MenuHandler) UpdateMenu(c *gin.Context) {
	id := c.Param("id")

	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid menu ID"})
		return
	}

	var menu model.MenuItem
	if err := c.ShouldBindJSON(&menu); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	menu.ID, _ = primitive.ObjectIDFromHex(id)
	menu.BeforeUpdate()

	ctx := c.Request.Context()
	if err := h.service.Update(ctx, &menu); err != nil {
		h.log.Error("Failed to update menu", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update menu"})
		return
	}
	if h.producer != nil {
		h.producer.SendMenuUpdated(menu.ID.Hex(), menu.Name)
	}

	c.JSON(http.StatusOK, gin.H{"data": menu})
}

func (h *MenuHandler) DeleteMenu(c *gin.Context) {
	id := c.Param("id")

	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid menu ID"})
		return
	}

	ctx := c.Request.Context()
	if err := h.service.Delete(ctx, id); err != nil {
		h.log.Error("Failed to delete menu", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete menu"})
		return
	}
	if h.producer != nil {
		h.producer.SendMenuDeleted(id)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Menu deleted"})
}

func (h *MenuHandler) SearchMenus(c *gin.Context) {
	tagsParam := c.Query("tags")
	if tagsParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tags parameter is required"})
		return
	}

	tags := strings.Split(tagsParam, ",")

	cleanTags := []string{}
	for _, tag := range tags {
		if strings.TrimSpace(tag) != "" {
			cleanTags = append(cleanTags, strings.TrimSpace(tag))
		}
	}

	ctx := c.Request.Context()
	menus, err := h.service.SearchByTags(ctx, cleanTags)
	if err != nil {
		h.log.Error("Failed to search menus", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search menus"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": menus})
}
