package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MenuItem struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string             `bson:"name" json:"name" validate:"required"`
	Description string             `bson:"description" json:"description"`
	Category    string             `bson:"category" json:"category" validate:"required"`
	Ingredients []string           `bson:"ingredients" json:"ingredients"`
	KBJU        KBJU               `bson:"kbju" json:"kbju"`
	Tags        []string           `bson:"tags" json:"tags"`
	ImageURL    string             `bson:"image_url" json:"image_url"`
	IsAvailable bool               `bson:"is_available" json:"is_available"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type KBJU struct {
	Calories float64 `bson:"calories" json:"calories"` // на 100г
	Proteins float64 `bson:"proteins" json:"proteins"`
	Fats     float64 `bson:"fats" json:"fats"`
	Carbs    float64 `bson:"carbs" json:"carbs"`
}

func (m *MenuItem) BeforeCreate() {
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	if m.ID.IsZero() {
		m.ID = primitive.NewObjectID()
	}
}

func (m *MenuItem) BeforeUpdate() {
	m.UpdatedAt = time.Now()
}
