package config

import (
	"log"
	"os"
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
			Port:            5432,
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			DBName:          getEnv("DB_NAME", "user_db"),
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: 10 * time.Minute,
		},
		JWT: JWTConfig{
			SecretKey:       getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
		Kafka: KafkaConfig{
			Brokers: []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
