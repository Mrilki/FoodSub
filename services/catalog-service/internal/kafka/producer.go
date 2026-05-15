package kafka

import (
	"encoding/json"
	"fmt"

	"github.com/Mrilki/catalog-service/pkg/logger"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"
)

type KafkaProducer interface {
	SendMenuCreated(menuID, name, category string) error
	SendMenuUpdated(menuID, name string) error
	SendMenuDeleted(menuID string) error
	Close()
}

type kafkaProducer struct {
	producer *kafka.Producer
	topic    string
	log      *logger.Logger
}

func NewKafkaProducer(brokers string, topic string, log *logger.Logger) (KafkaProducer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	log.Info("Kafka producer initialized", zap.String("topic", topic))

	return &kafkaProducer{
		producer: p,
		topic:    topic,
		log:      log,
	}, nil
}

func (p *kafkaProducer) SendMenuCreated(menuID, name, category string) error {
	event := map[string]interface{}{
		"event_type": "menu.created",
		"menu_id":    menuID,
		"name":       name,
		"category":   category,
	}

	return p.sendEvent(event)
}

func (p *kafkaProducer) SendMenuUpdated(menuID, name string) error {
	event := map[string]interface{}{
		"event_type": "menu.updated",
		"menu_id":    menuID,
		"name":       name,
	}

	return p.sendEvent(event)
}

func (p *kafkaProducer) SendMenuDeleted(menuID string) error {
	event := map[string]interface{}{
		"event_type": "menu.deleted",
		"menu_id":    menuID,
	}

	return p.sendEvent(event)
}

func (p *kafkaProducer) sendEvent(event map[string]interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		p.log.Error("Failed to marshal event", zap.Error(err))
		return err
	}

	p.log.Debug("Sending Kafka event", zap.Any("event", event))

	err = p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &p.topic,
			Partition: kafka.PartitionAny,
		},
		Value: data,
	}, nil)

	if err != nil {
		p.log.Error("Failed to produce Kafka message", zap.Error(err))
		return err
	}

	return nil
}

func (p *kafkaProducer) Close() {
	p.producer.Flush(15 * 1000)
	p.producer.Close()
	p.log.Info("Kafka producer closed")
}
