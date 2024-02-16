package confluentkafka

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"os"
)

type ProducerService struct {
	producer *kafka.Producer
	topic    string
}

func NewProducerService(topic string) (*ProducerService, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("KAFKA_BOOTSTRAP_SERVERS"),
		"security.protocol": "SASL_SSL",
		"sasl.mechanisms":   "PLAIN",
		"sasl.username":     os.Getenv("KAFKA_USERNAME"),
		"sasl.password":     os.Getenv("KAFKA_PASSWORD"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %v", err)
	}

	return &ProducerService{
		producer: producer,
		topic:    topic,
	}, nil
}

func (p *ProducerService) ProduceMessage(message string) error {
	err := p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &p.topic, Partition: kafka.PartitionAny},
		Value:          []byte(message),
	}, nil)

	if err != nil {
		return fmt.Errorf("failed to produce message: %v", err)
	}

	p.producer.Flush(15 * 1000)

	return nil
}

func (p *ProducerService) Close() {
	p.producer.Close()
}
