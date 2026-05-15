package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Mrilki/catalog-service/pkg/logger"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"
)

type EventHandler func(event map[string]interface{}) error

type KafkaConsumer interface {
	Start(ctx context.Context) error
	Stop()
	RegisterHandler(eventType string, handler EventHandler)
}

type kafkaConsumer struct {
	consumer      *kafka.Consumer
	topics        []string
	eventHandlers map[string]EventHandler
	log           *logger.Logger
	running       bool
}

func NewKafkaConsumer(
	brokers string,
	groupID string,
	topics []string,
	log *logger.Logger,
) (KafkaConsumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  brokers,
		"group.id":           groupID,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	log.Info("Kafka consumer initialized", zap.Strings("topics", topics))

	return &kafkaConsumer{
		consumer:      c,
		topics:        topics,
		eventHandlers: make(map[string]EventHandler),
		log:           log,
		running:       false,
	}, nil
}

func (c *kafkaConsumer) RegisterHandler(eventType string, handler EventHandler) {
	c.eventHandlers[eventType] = handler
	c.log.Debug("Registered handler for event", zap.String("type", eventType))
}

func (c *kafkaConsumer) Start(ctx context.Context) error {
	if err := c.consumer.SubscribeTopics(c.topics, nil); err != nil {
		return fmt.Errorf("failed to subscribe to topics: %w", err)
	}

	c.running = true
	c.log.Info("Kafka consumer started", zap.Strings("topics", c.topics))

	go func() {
		for c.running {
			select {
			case <-ctx.Done():
				c.log.Info("Context cancelled, stopping consumer")
				return
			default:
				msg, err := c.consumer.ReadMessage(100 * time.Millisecond)
				if err != nil {
					continue // Таймаут — нормально
				}
				c.processMessage(msg)
			}
		}
	}()

	return nil
}

func (c *kafkaConsumer) processMessage(msg *kafka.Message) {
	var event map[string]interface{}
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		c.log.Error("Failed to unmarshal Kafka message", zap.Error(err))
		return
	}

	eventType, ok := event["event_type"].(string)
	if !ok {
		c.log.Warn("Event type not found in message")
		return
	}

	c.log.Info("Received Kafka event",
		zap.String("type", eventType),
		zap.String("topic", *msg.TopicPartition.Topic))

	if handler, exists := c.eventHandlers[eventType]; exists {
		if err := handler(event); err != nil {
			c.log.Error("Event handler failed",
				zap.String("type", eventType),
				zap.Error(err))
		}
	}
}

func (c *kafkaConsumer) Stop() {
	c.running = false
	c.consumer.Close()
	c.log.Info("Kafka consumer stopped")
}
