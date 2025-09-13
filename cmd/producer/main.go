package main

import (
	"Order-tracker-service/config"
	"Order-tracker-service/internal/domain"
	"Order-tracker-service/internal/transport/kafka"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

// Producer представляет Kafka producer
type Producer struct {
	producer *kafka.Producer
	config   *config.KafkaConfig
}

// NewProducer создает новый экземпляр producer
func NewProducer(cfg *config.KafkaConfig) (*Producer, error) {
	prod, err := kafka.NewProducer(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return &Producer{
		producer: prod,
		config:   cfg,
	}, nil
}

// generateOrder генерирует заказ с уникальным ID
func generateOrder() *domain.Order {
	// Генерируем уникальный ID
	orderUID := uuid.New().String()

	// Генерируем случайные данные
	smID := rand.Intn(100) + 1
	chrtID := rand.Intn(10000) + 1000
	price := rand.Intn(5000) + 1000
	amount := price + rand.Intn(1000)

	// Генерируем случайные имена и адреса
	names := []string{"Иван Иванов", "Петр Петров", "Мария Сидорова", "Анна Козлова", "Дмитрий Волков"}
	cities := []string{"Москва", "Санкт-Петербург", "Новосибирск", "Екатеринбург", "Казань"}
	streets := []string{"ул. Ленина", "пр. Мира", "ул. Пушкина", "ул. Гагарина", "ул. Советская"}

	selectedName := names[rand.Intn(len(names))]
	selectedCity := cities[rand.Intn(len(cities))]
	selectedStreet := streets[rand.Intn(len(streets))]

	// Генерируем случайные товары
	items := []domain.Item{
		{
			ChrtID:      chrtID,
			TrackNumber: fmt.Sprintf("WBIL%d", rand.Intn(10000)),
			Price:       price,
			RID:         uuid.New().String(),
			Name:        "Смартфон",
			Sale:        rand.Intn(20),
			Size:        "M",
			TotalPrice:  price - (price * rand.Intn(20) / 100),
			NmID:        rand.Intn(100000) + 10000,
			Brand:       "Samsung",
			Status:      202,
		},
		{
			ChrtID:      chrtID + 1,
			TrackNumber: fmt.Sprintf("WBIL%d", rand.Intn(10000)),
			Price:       price + 500,
			RID:         uuid.New().String(),
			Name:        "Наушники",
			Sale:        rand.Intn(15),
			Size:        "S",
			TotalPrice:  price + 500 - ((price + 500) * rand.Intn(15) / 100),
			NmID:        rand.Intn(100000) + 20000,
			Brand:       "Apple",
			Status:      202,
		},
	}

	return &domain.Order{
		OrderUID:          orderUID,
		TrackNumber:       fmt.Sprintf("WBILMTESTTRACK%d", rand.Intn(1000)),
		Entry:             "WBIL",
		Locale:            "ru",
		InternalSignature: "",
		CustomerID:        fmt.Sprintf("customer_%d", rand.Intn(1000)),
		DeliveryService:   "meest",
		ShardKey:          strconv.Itoa(rand.Intn(10) + 1),
		SmID:              smID,
		DateCreated:       time.Now(),
		OofShard:          "1",

		Delivery: domain.Delivery{
			Name:    selectedName,
			Phone:   fmt.Sprintf("+7%d", rand.Intn(9000000000)+1000000000),
			Zip:     fmt.Sprintf("%d", rand.Intn(999999)+100000),
			City:    selectedCity,
			Address: fmt.Sprintf("%s, д. %d", selectedStreet, rand.Intn(100)+1),
			Region:  "Московская область",
			Email:   fmt.Sprintf("test%d@example.com", rand.Intn(1000)),
		},

		Payment: domain.Payment{
			Transaction:  uuid.New().String(),
			RequestID:    uuid.New().String(),
			Currency:     "RUB",
			Provider:     "wbpay",
			Amount:       amount,
			PaymentDT:    time.Now().Unix(),
			Bank:         "alpha",
			DeliveryCost: 150,
			GoodsTotal:   amount - 150,
			CustomFee:    0,
		},

		Items: items,
	}
}

// sendOrder отправляет заказ в Kafka
func (p *Producer) sendOrder(order *domain.Order) error {
	// Сериализуем заказ в JSON
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	// Отправляем сообщение в Kafka
	if err := p.producer.SendMessage(p.config.Topic, orderJSON); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// startGeneration запускает генерацию и отправку заказов
func (p *Producer) startGeneration(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Starting order generation every %v", interval)

	for range ticker.C {
		order := generateOrder()

		if err := p.sendOrder(order); err != nil {
			log.Printf("Failed to send order %s: %v", order.OrderUID, err)
		} else {
			log.Printf("Order sent successfully: %s (Customer: %s, Amount: %d RUB)",
				order.OrderUID, order.CustomerID, order.Payment.Amount)
		}
	}
}

func main() {
	// Инициализируем генератор случайных чисел
	rand.Seed(time.Now().UnixNano())

	// Загружаем переменные окружения
	if err := godotenv.Load("/Users/veraryabova/Desktop/Go/Order-tracker-service/.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Создаем producer
	producer, err := NewProducer(&cfg.Kafka)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}

	// Получаем интервал генерации из переменной окружения (по умолчанию 5 секунд)
	intervalStr := os.Getenv("GENERATION_INTERVAL")
	interval := 5 * time.Second
	if intervalStr != "" {
		if parsedInterval, err := time.ParseDuration(intervalStr); err == nil {
			interval = parsedInterval
		}
	}

	log.Printf("Kafka Producer Configuration:")
	log.Printf("   Brokers: %v", cfg.Kafka.Brokers)
	log.Printf("   Topic: %s", cfg.Kafka.Topic)
	log.Printf("   Generation Interval: %v", interval)

	// Обработка сигналов для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем генерацию в отдельной горутине
	go producer.startGeneration(interval)

	// Ожидаем сигнал завершения
	<-sigChan
	log.Println("Received shutdown signal, stopping producer...")

	// Закрываем producer
	if err := producer.producer.Close(); err != nil {
		log.Printf("Error closing producer: %v", err)
	}

	log.Println("Producer shutdown completed")
}
