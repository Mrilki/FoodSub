package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Mrilki/catalog-service/internal/model"
	"github.com/Mrilki/catalog-service/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type MenuRepository interface {
	Create(ctx context.Context, menu *model.MenuItem) error
	GetAll(ctx context.Context) ([]*model.MenuItem, error)
	GetByID(ctx context.Context, id string) (*model.MenuItem, error)
	Update(ctx context.Context, menu *model.MenuItem) error
	Delete(ctx context.Context, id string) error
	SearchByTags(ctx context.Context, tags []string) ([]*model.MenuItem, error)
}

type mongoMenuRepository struct {
	collection *mongo.Collection
	database   string
	log        *logger.Logger
}

func NewMongoMenuRepository(client *mongo.Client, database string, log *logger.Logger) MenuRepository {
	collection := client.Database(database).Collection("menu_items")
	return &mongoMenuRepository{
		collection: collection,
		database:   database,
		log:        log,
	}
}

func (r *mongoMenuRepository) Create(ctx context.Context, menu *model.MenuItem) error {
	menu.BeforeCreate()

	_, err := r.collection.InsertOne(ctx, menu)
	if err != nil {
		r.log.Error("failed to create menu",
			zap.String("name", menu.Name),
			zap.Error(err))
		return fmt.Errorf("failed to create menu: %w", err)
	}
	r.log.Info("create menu",
		zap.String("id", menu.ID.Hex()),
		zap.String("name", menu.Name))
	return nil
}

func (r *mongoMenuRepository) GetAll(ctx context.Context) ([]*model.MenuItem, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to get menus: %w", err)
	}
	defer cursor.Close(ctx)
	var menus []*model.MenuItem
	if err := cursor.All(ctx, &menus); err != nil {
		return nil, fmt.Errorf("failed to decode menus: %w", err)
	}
	return menus, nil
}

func (r *mongoMenuRepository) GetByID(ctx context.Context, id string) (*model.MenuItem, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id format: %w", err)
	}

	var menu model.MenuItem
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&menu)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("menu not found")
		}
		return nil, fmt.Errorf("failed to get menu: %w", err)
	}

	return &menu, nil
}

func (r *mongoMenuRepository) Update(ctx context.Context, menu *model.MenuItem) error {
	menu.BeforeUpdate()

	filter := bson.M{"_id": menu.ID}

	update := bson.M{"$set": menu}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update menu: %w", err)
	}

	r.log.Info("update menu", zap.String("id", menu.ID.Hex()))
	return nil
}

func (r *mongoMenuRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid id format: %w", err)
	}
	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return fmt.Errorf("failed to delete menu: %w", err)
	}
	r.log.Info("delete menu", zap.String("id", id))
	return nil
}

func (r *mongoMenuRepository) SearchByTags(ctx context.Context, tags []string) ([]*model.MenuItem, error) {
	filter := bson.M{"tags": bson.M{"$in": tags}}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find menus: %w", err)
	}
	defer cursor.Close(ctx)
	var menus []*model.MenuItem
	if err = cursor.All(ctx, &menus); err != nil {
		return nil, fmt.Errorf("failed to decode menus: %w", err)
	}
	return menus, nil
}
