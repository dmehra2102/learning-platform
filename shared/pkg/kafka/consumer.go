package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type MessageHandler func(ctx context.Context, key, value []byte) error

type Consumer struct {
	reader  *kafka.Reader
	handler MessageHandler
	logger  *zap.Logger
}

func NewConsumer(brokers []string, topic, groupID string, handler MessageHandler, logger *zap.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:          brokers,
		Topic:            topic,
		GroupID:          groupID,
		MinBytes:         10e3, // 10KB
		MaxBytes:         10e6, // 10MB
		CommitInterval:   1,
		StartOffset:      kafka.LastOffset,
		MaxAttempts:      3,
		SessionTimeout:   10,
		RebalanceTimeout: 10,
	})

	return &Consumer{
		reader:  reader,
		handler: handler,
		logger:  logger,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	c.logger.Info("starting kafka consumer",
		zap.String("topic", c.reader.Config().Topic),
		zap.String("group_id", c.reader.Config().GroupID),
	)

	for {
		select {
		case <-ctx.Done():
			return c.Close()
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				c.logger.Error("failed to fetch message", zap.Error(err))
				continue
			}

			if err := c.handler(ctx, msg.Key, msg.Value); err != nil {
				c.logger.Error("failed to handle message",
					zap.Error(err),
					zap.String("key", string(msg.Key)),
				)
				continue
			}

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				c.logger.Error("failed to commit message", zap.Error(err))
			}
		}
	}
}

func (c *Consumer) Close() error {
	c.logger.Info("closing kafka consumer")
	return c.reader.Close()
}

func UnmarshalMessage(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
