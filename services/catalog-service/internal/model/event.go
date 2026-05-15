package model

import "time"

type KafkaEvent struct {
	EventType string                 `json:"event_type"`
	TraceID   string                 `json:"trace_id"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
}

type OrderScheduledEvent struct {
	OrderID        string    `json:"order_id"`
	SubscriptionID string    `json:"subscription_id"`
	UserID         string    `json:"user_id"`
	MenuItemIDs    []string  `json:"menu_item_ids"`
	DeliveryDate   time.Time `json:"delivery_date"`
}

type OrderDeliveredEvent struct {
	OrderID     string    `json:"order_id"`
	DeliveredAt time.Time `json:"delivered_at"`
	Rating      int       `json:"rating,omitempty"`
}

type MenuCreatedEvent struct {
	MenuID    string    `json:"menu_id"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
}

type MenuUpdatedEvent struct {
	MenuID    string    `json:"menu_id"`
	Name      string    `json:"name"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MenuDeletedEvent struct {
	MenuID    string    `json:"menu_id"`
	DeletedAt time.Time `json:"deleted_at"`
}
