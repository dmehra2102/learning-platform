package config

import (
	"os"
	"strconv"
	"time"

	"github.com/dmehra2102/learning-platform/shared/pkg/database"
)

type Config struct {
	Server   ServerConfig
	Database database.Config
	JWT      JWTConfig
	Kafka    KafkaConfig
	App      AppConfig
}

type ServerConfig struct {
	Port int
}

type JWTConfig struct {
	SecretKey           string
	AccessTokenExpiry   time.Duration
	RefreshTokenExpiry  time.Duration
}

type KafkaConfig struct {
	Brokers []string
}

type AppConfig struct {
	Environment string
	LogLevel    string
}

func Load() Config {
	return Config{
		Server: ServerConfig{
			Port: getEnvInt("SERVER_PORT", 50052),
		},
		Database: database.Config{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			DBName:          getEnv("DB_NAME", "course_db"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: time.Duration(getEnvInt("DB_CONN_MAX_LIFETIME", 5)) * time.Minute,
			ConnMaxIdleTime: time.Duration(getEnvInt("DB_CONN_MAX_IDLE_TIME", 10)) * time.Minute,
		},
		JWT: JWTConfig{
			SecretKey:          getEnv("JWT_SECRET", "secret_key"),
			AccessTokenExpiry:  time.Duration(getEnvInt("JWT_ACCESS_EXPIRY_MIN", 15)) * time.Minute,
			RefreshTokenExpiry: time.Duration(getEnvInt("JWT_REFRESH_EXPIRY_DAYS", 7)) * 24 * time.Hour,
		},
		Kafka: KafkaConfig{
			Brokers: []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		},
		App: AppConfig{
			Environment: getEnv("APP_ENV", "development"),
			LogLevel:    getEnv("LOG_LEVEL", "info"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
