package service

import (
	"context"

	"github.com/Mrilki/catalog-service/internal/middleware"
	"github.com/Mrilki/catalog-service/internal/model"
	"github.com/Mrilki/catalog-service/internal/repository"
	"github.com/Mrilki/catalog-service/pkg/logger"
	"go.uber.org/zap"
)

type MenuService interface {
	GetAll(ctx context.Context) ([]*model.MenuItem, error)
	GetByID(ctx context.Context, id string) (*model.MenuItem, error)
	Create(ctx context.Context, menu *model.MenuItem) error
	Update(ctx context.Context, menu *model.MenuItem) error
	Delete(ctx context.Context, id string) error
	SearchByTags(ctx context.Context, tags []string) ([]*model.MenuItem, error)
	ProcessOrderScheduled(event map[string]interface{}) error
	ProcessOrderDelivered(event map[string]interface{}) error
}

type menuService struct {
	repo  repository.MenuRepository
	cache repository.RedisRepository
	log   *logger.Logger
}

func NewMenuService(
	repo repository.MenuRepository,
	cache repository.RedisRepository,
	log *logger.Logger,
) MenuService {
	return &menuService{
		repo:  repo,
		cache: cache,
		log:   log,
	}
}

func (s *menuService) GetAll(ctx context.Context) ([]*model.MenuItem, error) {
	cacheKey := "menu:all"

	cached, err := s.cache.GetMenu(ctx, cacheKey)
	if err != nil {
		s.log.Error("Cache get failed", zap.Error(err))
	}
	if cached != nil {
		s.log.Debug("Returning from cache")
		middleware.IncCacheHit()
		return cached, nil
	}

	middleware.IncCacheMiss()

	s.log.Debug("Cache miss, querying MongoDB")
	menus, err := s.repo.GetAll(ctx)
	if err != nil {
		s.log.Error("Failed to get menus from repository", zap.Error(err))
		return nil, err
	}

	if err := s.cache.SetMenu(ctx, cacheKey, menus); err != nil {
		s.log.Warn("Failed to cache menus", zap.Error(err))
	}

	return menus, nil
}

func (s *menuService) GetByID(ctx context.Context, id string) (*model.MenuItem, error) {
	cached, err := s.cache.GetMenuItem(ctx, id)
	if err != nil {
		s.log.Error("Cache get failed", zap.Error(err))
	}
	if cached != nil {
		s.log.Debug("Returning item from cache")
		return cached, nil
	}

	s.log.Debug("Cache miss, querying MongoDB")
	menu, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Warn("Menu not found", zap.String("id", id), zap.Error(err))
		return nil, err
	}

	if err := s.cache.SetMenuItem(ctx, id, menu); err != nil {
		s.log.Warn("Failed to cache menu item", zap.Error(err))
	}

	return menu, nil
}

func (s *menuService) Create(ctx context.Context, menu *model.MenuItem) error {
	s.log.Info("Creating new menu", zap.String("name", menu.Name))

	if menu.Name == "" {
		return &ValidationError{Field: "name", Message: "name is required"}
	}
	if menu.Category == "" {
		return &ValidationError{Field: "category", Message: "category is required"}
	}

	if err := s.repo.Create(ctx, menu); err != nil {
		s.log.Error("Failed to create menu in repository", zap.Error(err))
		return err
	}

	if err := s.cache.DeleteMenu(ctx, "menu:all"); err != nil {
		s.log.Warn("Failed to invalidate cache", zap.Error(err))
	}

	return nil
}

func (s *menuService) Update(ctx context.Context, menu *model.MenuItem) error {
	s.log.Info("Updating menu", zap.String("id", menu.ID.Hex()))

	if err := s.repo.Update(ctx, menu); err != nil {
		s.log.Error("Failed to update menu in repository", zap.Error(err))
		return err
	}

	if err := s.cache.DeleteMenu(ctx, "menu:all"); err != nil {
		s.log.Warn("Failed to invalidate menu list cache", zap.Error(err))
	}
	if err := s.cache.DeleteMenuItem(ctx, menu.ID.Hex()); err != nil {
		s.log.Warn("Failed to invalidate menu item cache", zap.Error(err))
	}

	return nil
}

func (s *menuService) Delete(ctx context.Context, id string) error {
	s.log.Info("Deleting menu", zap.String("id", id))

	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.Error("Failed to delete menu in repository", zap.Error(err))
		return err
	}

	if err := s.cache.DeleteMenu(ctx, "menu:all"); err != nil {
		s.log.Warn("Failed to invalidate menu list cache", zap.Error(err))
	}
	if err := s.cache.DeleteMenuItem(ctx, id); err != nil {
		s.log.Warn("Failed to invalidate menu item cache", zap.Error(err))
	}

	return nil
}

func (s *menuService) SearchByTags(ctx context.Context, tags []string) ([]*model.MenuItem, error) {
	s.log.Debug("Searching menus by tags", zap.Strings("tags", tags))
	menus, err := s.repo.SearchByTags(ctx, tags)
	if err != nil {
		s.log.Error("Failed to search menus", zap.Error(err))
		return nil, err
	}
	return menus, nil
}

func (s *menuService) ProcessOrderScheduled(event map[string]interface{}) error {
	s.log.Info("Processing order.scheduled event", zap.Any("event", event))

	menuItemIDs, ok := event["menu_item_ids"].([]interface{})
	if !ok {
		s.log.Warn("menu_item_ids not found in event")
		return nil
	}

	for _, id := range menuItemIDs {
		menuID, ok := id.(string)
		if !ok {
			continue
		}
		s.log.Debug("Incrementing popularity counter", zap.String("menu_id", menuID))
	}

	return nil
}

func (s *menuService) ProcessOrderDelivered(event map[string]interface{}) error {
	s.log.Info("Processing order.delivered event", zap.Any("event", event))
	return nil
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
