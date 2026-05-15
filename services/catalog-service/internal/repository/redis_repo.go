package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Mrilki/catalog-service/internal/model"
	"github.com/Mrilki/catalog-service/pkg/logger"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type RedisRepository interface {
	GetMenu(ctx context.Context, key string) ([]*model.MenuItem, error)
	SetMenu(ctx context.Context, key string, menus []*model.MenuItem) error
	DeleteMenu(ctx context.Context, key string) error
	GetMenuItem(ctx context.Context, id string) (*model.MenuItem, error)
	SetMenuItem(ctx context.Context, id string, menu *model.MenuItem) error
	DeleteMenuItem(ctx context.Context, id string) error
}

type redisRepository struct {
	client   *redis.Client
	cacheTTL time.Duration
	log      *logger.Logger
}

func NewRedisRepository(client *redis.Client, cacheTTL int, log *logger.Logger) RedisRepository {
	return &redisRepository{
		client:   client,
		cacheTTL: time.Duration(cacheTTL) * time.Second,
		log:      log,
	}
}

func (r *redisRepository) GetMenu(ctx context.Context, key string) ([]*model.MenuItem, error) {
	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Кэш не найден
	}
	if err != nil {
		r.log.Error("Failed to get from Redis", zap.String("key", key), zap.Error(err))
		return nil, err
	}

	var menus []*model.MenuItem
	if err := json.Unmarshal([]byte(data), &menus); err != nil {
		r.log.Error("Failed to unmarshal menu from Redis", zap.Error(err))
		return nil, err
	}

	r.log.Debug("Cache hit", zap.String("key", key))
	return menus, nil
}

func (r *redisRepository) SetMenu(ctx context.Context, key string, menus []*model.MenuItem) error {
	data, err := json.Marshal(menus)
	if err != nil {
		r.log.Error("Failed to marshal menu", zap.Error(err))
		return err
	}

	if err := r.client.Set(ctx, key, data, r.cacheTTL).Err(); err != nil {
		r.log.Error("Failed to set in Redis", zap.String("key", key), zap.Error(err))
		return err
	}

	r.log.Debug("Cache set", zap.String("key", key), zap.Duration("ttl", r.cacheTTL))
	return nil
}

func (r *redisRepository) DeleteMenu(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		r.log.Error("Failed to delete from Redis", zap.String("key", key), zap.Error(err))
		return err
	}

	r.log.Debug("Cache deleted", zap.String("key", key))
	return nil
}

func (r *redisRepository) GetMenuItem(ctx context.Context, id string) (*model.MenuItem, error) {
	key := fmt.Sprintf("menu:item:%s", id)
	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		r.log.Error("Failed to get menu item from Redis", zap.String("id", id), zap.Error(err))
		return nil, err
	}

	var menu model.MenuItem
	if err := json.Unmarshal([]byte(data), &menu); err != nil {
		r.log.Error("Failed to unmarshal menu item from Redis", zap.Error(err))
		return nil, err
	}

	r.log.Debug("Cache hit (item)", zap.String("id", id))
	return &menu, nil
}

func (r *redisRepository) SetMenuItem(ctx context.Context, id string, menu *model.MenuItem) error {
	key := fmt.Sprintf("menu:item:%s", id)
	data, err := json.Marshal(menu)
	if err != nil {
		r.log.Error("Failed to marshal menu item", zap.Error(err))
		return err
	}

	if err := r.client.Set(ctx, key, data, r.cacheTTL).Err(); err != nil {
		r.log.Error("Failed to set menu item in Redis", zap.String("id", id), zap.Error(err))
		return err
	}

	r.log.Debug("Cache set (item)", zap.String("id", id), zap.Duration("ttl", r.cacheTTL))
	return nil
}

func (r *redisRepository) DeleteMenuItem(ctx context.Context, id string) error {
	key := fmt.Sprintf("menu:item:%s", id)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		r.log.Error("Failed to delete menu item from Redis", zap.String("id", id), zap.Error(err))
		return err
	}

	r.log.Debug("Cache deleted (item)", zap.String("id", id))
	return nil
}
