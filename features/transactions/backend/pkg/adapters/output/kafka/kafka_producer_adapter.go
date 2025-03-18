package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/agmanchon/blockchainmonitor/features/transactions/backend/pkg/usecases"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaProducerAdapterOptions struct {
	Producer *kafka.Producer
	Topic    string
}

type KafkaProducerAdapter struct {
	producer *kafka.Producer
	topic    string
}

var _ usecases.KafkaProducer = (*KafkaProducerAdapter)(nil)

func NewKafkaProducerAdapter(options KafkaProducerAdapterOptions) (*KafkaProducerAdapter, error) {

	if options.Producer == nil {
		return nil, fmt.Errorf("kafka producer is required")
	}
	if options.Topic == "" {
		return nil, fmt.Errorf("kafka topic is required")
	}

	return &KafkaProducerAdapter{
		producer: options.Producer,
		topic:    options.Topic,
	}, nil
}

// SendEvent sends an event to the Kafka topic
func (k *KafkaProducerAdapter) SendEvent(ctx context.Context, event usecases.KafkaEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	deliveryChan := make(chan kafka.Event)
	defer close(deliveryChan)

	err = k.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &k.topic, Partition: kafka.PartitionAny},
		Value:          data,
	}, deliveryChan)

	if err != nil {
		return fmt.Errorf("failed to send event to Kafka: %w", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-deliveryChan:
		m := e.(*kafka.Message)
		if m.TopicPartition.Error != nil {
			return fmt.Errorf("failed to deliver message to Kafka: %w", m.TopicPartition.Error)
		}
	}

	return nil
}

// Close closes the Kafka producer
func (k *KafkaProducerAdapter) Close() {
	k.producer.Close()
}
