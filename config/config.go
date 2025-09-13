package config

import (
	"os"
	"strconv"
)

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	Kafka    KafkaConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
}

func LoadConfig() (*Config, error) {
	var config Config

	// Загружаем конфигурацию базы данных
	config.Database = DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvAsInt("DB_PORT", 5433),
		Username: getEnv("DB_USERNAME", "postgres"),
		Password: getEnv("DB_PASSWORD", "admin"),
		Database: getEnv("DB_DATABASE", "orders"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// Загружаем конфигурацию сервера
	config.Server = ServerConfig{
		Port: getEnv("SERVER_PORT", "8080"),
		Host: getEnv("SERVER_HOST", "localhost"),
	}

	// Загружаем конфигурацию Kafka
	config.Kafka = KafkaConfig{
		Brokers: getEnvAsSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
		Topic:   getEnv("KAFKA_TOPIC", "orders"),
	}

	return &config, nil
}

type DatabaseConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	SSLMode  string
}

// getEnv получает переменную окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt получает переменную окружения как int или возвращает значение по умолчанию
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsSlice получает переменную окружения как slice строк или возвращает значение по умолчанию
func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Простое разделение по запятой, можно улучшить для более сложных случаев
		values := []string{}
		for _, v := range splitString(value, ",") {
			if trimmed := trimString(v); trimmed != "" {
				values = append(values, trimmed)
			}
		}
		if len(values) > 0 {
			return values
		}
	}
	return defaultValue
}

// splitString разделяет строку по разделителю
func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}

	result := []string{}
	start := 0

	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}

	if start < len(s) {
		result = append(result, s[start:])
	}

	return result
}

// trimString удаляет пробелы в начале и конце строки
func trimString(s string) string {
	start := 0
	end := len(s)

	// Удаляем пробелы в начале
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	// Удаляем пробелы в конце
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}
