package interceptor

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type LoggingInterceptor struct {
	logger *zap.Logger
}

func NewLoggingInterceptor(logger *zap.Logger) *LoggingInterceptor {
	return &LoggingInterceptor{logger: logger}
}

func (i *LoggingInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		statusCode := status.Code(err)

		i.logger.Info("unary RPC",
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
			zap.String("status", statusCode.String()),
			zap.Error(err),
		)

		return resp, err
	}
}

func (i *LoggingInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()

		err := handler(srv, ss)

		duration := time.Since(start)
		statusCode := status.Code(err)

		i.logger.Info("stream RPC",
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
			zap.String("status", statusCode.String()),
			zap.Error(err),
		)

		return err
	}
}
