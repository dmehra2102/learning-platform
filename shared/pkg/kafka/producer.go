package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Producer struct {
	writer *kafka.Writer
	logger *zap.Logger
}

func NewProducer(brokers []string, topic string, logger *zap.Logger) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
		Compression:  kafka.Snappy,
	}

	return &Producer{
		writer: writer,
		logger: logger,
	}
}

func (p *Producer) PublishMessage(ctx context.Context, key string, value any) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(key),
		Value: valueBytes,
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		p.logger.Error("failed to publish message",
			zap.Error(err),
			zap.String("key", key),
		)
		return fmt.Errorf("failed to publish message : %w", err)
	}

	p.logger.Debug("message published successfully",
		zap.String("topic", p.writer.Topic),
		zap.String("key", key),
	)

	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
