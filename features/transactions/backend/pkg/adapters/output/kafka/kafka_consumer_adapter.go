package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/agmanchon/blockchainmonitor/features/transactions/backend/pkg/usecases"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaConsumerAdapterOptions struct {
	Consumer *kafka.Consumer
	Topic    string
}

type KafkaConsumerAdapter struct {
	consumer *kafka.Consumer
	topic    string
}

var _ usecases.KafkaConsumer = (*KafkaConsumerAdapter)(nil)

func NewKafkaConsumerAdapter(options KafkaConsumerAdapterOptions) (*KafkaConsumerAdapter, error) {

	if options.Consumer == nil {
		return nil, fmt.Errorf("kafka consumer is required")
	}
	if options.Topic == "" {
		return nil, fmt.Errorf("kafka topic is required")
	}

	return &KafkaConsumerAdapter{
		consumer: options.Consumer,
		topic:    options.Topic,
	}, nil
}

// StartConsumer listens to the Kafka topic and processes messages
func (k *KafkaConsumerAdapter) StartConsumer(ctx context.Context, messageHandler func(usecases.KafkaEvent) error) error {
	// Subscribe to the topic
	err := k.consumer.Subscribe(k.topic, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", k.topic, err)
	}

	// Loop and consume messages
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, err := k.consumer.ReadMessage(-1)
			if err != nil {
				return fmt.Errorf("failed to read message: %w", err)
			}

			var event usecases.KafkaEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				return fmt.Errorf("failed to unmarshal event: %w", err)
			}

			// Handle the event using the provided messageHandler function
			if err := messageHandler(event); err != nil {
				return fmt.Errorf("failed to handle message: %w", err)
			}
		}
	}
}

// Close closes the Kafka consumer
func (k *KafkaConsumerAdapter) Close() {
	k.consumer.Close()
}
