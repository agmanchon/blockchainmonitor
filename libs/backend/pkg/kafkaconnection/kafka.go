package kafkaconnection

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaConnection struct {
	Producer *kafka.Producer
	Consumer *kafka.Consumer
	Topic    string
}

func NewKafkaConnectionFromConfig(brokers string, topic string) (*KafkaConnection, error) {
	configMap := &kafka.ConfigMap{
		"bootstrap.servers": brokers, // Kafka broker list
		"acks":              "all",   // Set the ack level (optional)
	}

	// Create a new producer using the provided brokers and config
	producer, err := kafka.NewProducer(configMap)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"group.id":          "my-group",
		"auto.offset.reset": "earliest",
	})

	// Use NewKafkaProducerAdapter to return the adapter with the given topic
	return &KafkaConnection{Producer: producer, Consumer: consumer, Topic: topic}, nil

}
