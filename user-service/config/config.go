package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dmehra2102/learning-platform/shared/pkg/database"
	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database database.Config
	JWT      JWTConfig
	Kafka    KafkaConfig
}

type ServerConfig struct {
	Port int
}

type JWTConfig struct {
	SecretKey       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type KafkaConfig struct {
	Brokers []string
}

func Load() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	return Config{
		Server: ServerConfig{
			Port: 50051,
		},
		Database: database.Config{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getIntEnv("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			DBName:          getEnv("DB_NAME", "user_db"),
			SSLMode:         "disable",
			MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: 10 * time.Minute,
		},
		JWT: JWTConfig{
			SecretKey:       getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
		Kafka: KafkaConfig{
			Brokers: getSliceEnv("KAFKA_BROKERS", []string{"localhost:9092"}),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if result, err := strconv.Atoi(key); err == nil {
			return result
		}
	}
	return defaultValue
}

func getSliceEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}