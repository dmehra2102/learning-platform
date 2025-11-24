package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/metadata"
)

var globalLogger *zap.Logger

func init() {
	InitLogger("production")
}

func InitLogger(environment string) {
	var config zap.Config

	if environment == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	globalLogger = logger
}

func GetLogger() *zap.Logger {
	return globalLogger
}

func Info(msg string, fields ...zap.Field) {
	globalLogger.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	globalLogger.Error(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	globalLogger.Warn(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	globalLogger.Debug(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	globalLogger.Fatal(msg, fields...)
}

func WithTrace(ctx context.Context) zap.Field {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return zap.String("trace_id", "")
	}

	traceIDs := md.Get("x-trace-id")
	if len(traceIDs) == 0 {
		return zap.String("trace_id", "")
	}

	return zap.String("trace_id", traceIDs[0])
}

func Sync() {
	if globalLogger != nil {
		_ = globalLogger.Sync()
	}
}
