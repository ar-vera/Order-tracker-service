package db

import (
	"Order-tracker-service/config"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// InitDB инициализирует подключение к базе данных
func InitDB(cfg *config.DatabaseConfig) (*sqlx.DB, error) {
	// Формируем DSN строку для подключения к PostgreSQL
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Password,
		cfg.Database,
		cfg.SSLMode,
	)

	// Подключаемся к базе данных
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Настраиваем параметры пула соединений
	db.SetMaxOpenConns(25)                 // Максимальное количество открытых соединений
	db.SetMaxIdleConns(5)                  // Максимальное количество неактивных соединений
	db.SetConnMaxLifetime(5 * time.Minute) // Максимальное время жизни соединения

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Successfully connected to PostgreSQL database: %s@%s:%d/%s",
		cfg.Username, cfg.Host, cfg.Port, cfg.Database)

	return db, nil
}

// CloseDB закрывает подключение к базе данных
func CloseDB(db *sqlx.DB) error {
	if db == nil {
		return nil
	}

	if err := db.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	log.Println("Database connection closed successfully")
	return nil
}

// HealthCheck проверяет состояние подключения к базе данных
func HealthCheck(db *sqlx.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}
