package main

import (
	"Order-tracker-service/config"
	"Order-tracker-service/internal/db"
	"Order-tracker-service/internal/repository"
	"Order-tracker-service/internal/service"
	httptransport "Order-tracker-service/internal/transport/http"
	"Order-tracker-service/internal/transport/kafka"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Загружаем переменные окружения
	// В контейнере переменные приходят из окружения docker-compose, поэтому .env может отсутствовать
	_ = godotenv.Load()

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализируем подключение к базе данных
	dataBase, err := db.InitDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Запускаем миграции
	db.RunMigrations(dataBase, "migrations")
	defer func() {
		if err := db.CloseDB(dataBase); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Создаем репозиторий
	repo := repository.NewOrderRepository(dataBase)

	// Создаем сервис
	orderService := service.NewOrderService(repo, 0)

	// Инициализируем HTTP хэндлер
	httpHandler := httptransport.NewHandler(orderService)
	router := httpHandler.InitRoutes()

	// Создаем HTTP сервер
	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	// Инициализируем Kafka консьюмер
	consumer, err := kafka.NewConsumer(&cfg.Kafka, orderService)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}

	// Запускаем консьюмер
	if err := consumer.Start(); err != nil {
		log.Fatalf("Failed to start Kafka consumer: %v", err)
	}

	// Запускаем HTTP сервер в отдельной горутине
	go func() {
		log.Printf("HTTP server starting on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	log.Printf("Order service, Kafka consumer and HTTP server initialized successfully")

	// Ожидаем сигнал для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Блокируемся до получения сигнала
	<-sigChan
	log.Println("Received shutdown signal, stopping services...")

	// Graceful shutdown HTTP сервера
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down HTTP server: %v", err)
	}

	// Останавливаем консьюмер
	if err := consumer.Stop(); err != nil {
		log.Printf("Error stopping Kafka consumer: %v", err)
	}

	log.Println("Application shutdown completed")
}
