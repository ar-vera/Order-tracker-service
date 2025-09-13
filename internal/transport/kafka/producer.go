package kafka

import (
	"Order-tracker-service/config"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
)

// Producer представляет Kafka producer
type Producer struct {
	producer sarama.SyncProducer
	config   *config.KafkaConfig
}

// NewProducer создает новый экземпляр producer
func NewProducer(cfg *config.KafkaConfig) (*Producer, error) {
	// Настройка конфигурации Sarama
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = 3
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Timeout = 10 * time.Second
	saramaConfig.Producer.Compression = sarama.CompressionSnappy

	// Создание producer
	producer, err := sarama.NewSyncProducer(cfg.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	log.Printf("Kafka producer created successfully, brokers: %v", cfg.Brokers)

	return &Producer{
		producer: producer,
		config:   cfg,
	}, nil
}

// SendMessage отправляет сообщение в Kafka
func (p *Producer) SendMessage(topic string, message []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(message),
		Key:   sarama.StringEncoder("order"),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	log.Printf("Message sent to topic %s, partition %d, offset %d", topic, partition, offset)
	return nil
}

// Close закрывает producer
func (p *Producer) Close() error {
	if err := p.producer.Close(); err != nil {
		return fmt.Errorf("failed to close producer: %w", err)
	}

	log.Println("Kafka producer closed successfully")
	return nil
}
