package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/dmehra2102/learning-platform/shared/pkg/database"
	"github.com/dmehra2102/learning-platform/shared/pkg/interceptor"
	"github.com/dmehra2102/learning-platform/shared/pkg/jwt"
	"github.com/dmehra2102/learning-platform/shared/pkg/kafka"
	"github.com/dmehra2102/learning-platform/shared/pkg/logger"
	pb "github.com/dmehra2102/learning-platform/shared/proto/user"
	"github.com/dmehra2102/learning-platform/user-service/config"
	"github.com/dmehra2102/learning-platform/user-service/internal/grpc"
	"github.com/dmehra2102/learning-platform/user-service/internal/repository"
	"github.com/dmehra2102/learning-platform/user-service/internal/service"
	"go.uber.org/zap"
	grpcLib "google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	// Initializing Logger
	logger.InitLogger("production")
	log := logger.GetLogger()
	defer logger.Sync()

	log.Info("starting user service")

	// Load configuration
	cfg := config.Load()

	// Initialize Database
	db, err := database.NewPostgresDB(cfg.Database, log)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Run migrations
	if err := runDBMigrations(db, log); err != nil {
		log.Fatal("failed to run migrations", zap.Error(err))
	}

	// Initialize JWT Manager
	jwtManager := jwt.NewManager(
		cfg.JWT.SecretKey,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
	)

	// Initialize Kafka producer
	kafkaProducer := kafka.NewProducer(
		cfg.Kafka.Brokers,
		kafka.TopicUserRegistered,
		log,
	)
	defer kafkaProducer.Close()

	// Initialize repository
	userRepo := repository.NewUserRepository(db)

	// Initialize Service
	userServer := service.NewUserService(userRepo, jwtManager, kafkaProducer, log)

	// Initialize gRPC server
	authInterceptor := interceptor.NewAuthInterceptor(jwtManager)
	loggingInterceptor := interceptor.NewLoggingInterceptor(log)
	recoveryInterceptor := interceptor.NewRecoveryInterceptor(log)

	grpcServer := grpcLib.NewServer(
		grpcLib.ChainUnaryInterceptor(
			recoveryInterceptor.Unary(),
			loggingInterceptor.Unary(),
			authInterceptor.Unary(),
		),
		grpcLib.ChainStreamInterceptor(
			recoveryInterceptor.Stream(),
			loggingInterceptor.Stream(),
			authInterceptor.Stream(),
		),
	)

	// Register services
	userHandler := grpc.NewUserHandler(userServer)
	pb.RegisterUserServiceServer(grpcServer, userHandler)

	// Register health check
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("user-service", grpc_health_v1.HealthCheckResponse_SERVING)

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	go func() {
		log.Info("user service listening", zap.Int("port", cfg.Server.Port))
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("failed to serve", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down user service")
	grpcServer.GracefulStop()
}

func runDBMigrations(db *database.DB, log *zap.Logger) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(255) UNQIUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			first_name VARCHAR(100) NOT NULL,
			last_name VARCHAR(100) NOT NULL,
			role VARCHAR(20) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
			avatar_url TEXT,
			bio TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_users_role ON users(role)`,
		`CREATE INDEX IF NOT EXISTS idx_users_status ON users(status)`,
	}

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	log.Info("Database migrations completed successfully")
	return nil
}
