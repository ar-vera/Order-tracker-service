package kafka

import (
	"Order-tracker-service/config"
	"Order-tracker-service/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

// Consumer представляет Kafka консьюмер
type Consumer struct {
	config    *config.KafkaConfig
	consumer  sarama.ConsumerGroup
	handler   MessageHandler
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	isRunning bool
	mu        sync.RWMutex
}

// MessageHandler интерфейс для обработки сообщений
type MessageHandler interface {
	HandleOrder(ctx context.Context, order *domain.Order) error
}

// NewConsumer создает новый экземпляр консьюмера
func NewConsumer(cfg *config.KafkaConfig, handler MessageHandler) (*Consumer, error) {
	// Настройка конфигурации Sarama
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	saramaConfig.Consumer.Group.Session.Timeout = 10 * time.Second
	saramaConfig.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	saramaConfig.Consumer.MaxProcessingTime = 500 * time.Millisecond
	saramaConfig.Consumer.Return.Errors = true

	// Создание консьюмера
	consumer, err := sarama.NewConsumerGroup(cfg.Brokers, "order-tracker-group", saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Consumer{
		config:   cfg,
		consumer: consumer,
		handler:  handler,
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

// Start запускает консьюмер
func (c *Consumer) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isRunning {
		return fmt.Errorf("consumer is already running")
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.consume()
	}()

	c.isRunning = true
	log.Printf("Kafka consumer started, listening to topic: %s", c.config.Topic)
	return nil
}

// Stop останавливает консьюмер
func (c *Consumer) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isRunning {
		return nil
	}

	c.cancel()
	c.wg.Wait()

	if err := c.consumer.Close(); err != nil {
		return fmt.Errorf("failed to close consumer: %w", err)
	}

	c.isRunning = false
	log.Println("Kafka consumer stopped")
	return nil
}

// consume основной цикл потребления сообщений
func (c *Consumer) consume() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			// Потребление сообщений
			topics := []string{c.config.Topic}
			err := c.consumer.Consume(c.ctx, topics, c)
			if err != nil {
				log.Printf("Error consuming messages: %v", err)
				time.Sleep(time.Second)
			}
		}
	}
}

// Setup вызывается в начале новой сессии
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	log.Println("Kafka consumer session started")
	return nil
}

// Cleanup вызывается в конце сессии
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	log.Println("Kafka consumer session ended")
	return nil
}

// ConsumeClaim обрабатывает сообщения из партиции
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			// Обработка сообщения
			if err := c.processMessage(message); err != nil {
				log.Printf("Error processing message: %v", err)
				// В реальном приложении здесь может быть логика retry или dead letter queue
			}

			// Подтверждение обработки сообщения
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

// processMessage обрабатывает отдельное сообщение
func (c *Consumer) processMessage(message *sarama.ConsumerMessage) error {
	log.Printf("Received message from topic %s, partition %d, offset %d",
		message.Topic, message.Partition, message.Offset)

	// Десериализация сообщения в структуру Order
	var order domain.Order
	if err := json.Unmarshal(message.Value, &order); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Валидация заказа
	if order.OrderUID == "" {
		return fmt.Errorf("invalid order: missing OrderUID")
	}

	// Обработка заказа через handler
	if err := c.handler.HandleOrder(c.ctx, &order); err != nil {
		return fmt.Errorf("failed to handle order %s: %w", order.OrderUID, err)
	}

	log.Printf("Successfully processed order: %s", order.OrderUID)
	return nil
}

// IsRunning возвращает статус консьюмера
func (c *Consumer) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isRunning
}
