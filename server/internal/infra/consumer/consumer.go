package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"tcp-message-processor/common/pkg/closer"
	"tcp-message-processor/common/pkg/logger"
	"tcp-message-processor/internal/config"
	"tcp-message-processor/internal/events"
	"tcp-message-processor/pkg/broker"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type (
	Repository interface {
		Increment(ctx context.Context, username string, timestamp time.Time) error
	}

	Consumer struct {
		conn     *amqp.Connection
		channel  *amqp.Channel
		queue    string
		store    Repository
		stopChan chan struct{}
		doneChan chan struct{}
	}
)

func New(cfg config.RabbitMQConfig, store Repository) (*Consumer, error) {
	setup, err := broker.Consumer(broker.ConsumerConfig{
		URL:        cfg.URL,
		Exchange:   cfg.Exchange,
		Queue:      cfg.Queue,
		RoutingKey: "submission.created",
	})
	if err != nil {
		return nil, err
	}

	logger.Info("message broker consumer initialized", zap.String("queue", cfg.Queue))

	return &Consumer{
		conn:     setup.Conn,
		channel:  setup.Channel,
		queue:    setup.QueueName,
		store:    store,
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
	}, nil
}

func (c Consumer) Start(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		c.queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go c.runConsumer(ctx, msgs)
	return nil
}

func (c Consumer) runConsumer(ctx context.Context, msgs <-chan amqp.Delivery) {
	defer func() {
		logger.Info("consumer goroutine exiting, closing done channel")
		close(c.doneChan)
	}()

	logger.Info("consumer loop started")
	for {
		select {
		case <-ctx.Done():
			logger.Info("consumer stopped - context cancelled")
			return
		case <-c.stopChan:
			logger.Info("consumer stopped - stop signal received")
			return
		case msg, ok := <-msgs:
			if !ok {
				logger.Warn("consumer channel closed")
				return
			}
			c.handleMessage(ctx, msg)
		}
	}
}

func (c Consumer) handleMessage(ctx context.Context, msg amqp.Delivery) {
	if err := c.processMessage(ctx, msg); err != nil {
		logger.Error("failed to process message", zap.Error(err))
		nackMessage(msg)
		return
	}

	ackMessage(msg)
}

func (c Consumer) processMessage(ctx context.Context, msg amqp.Delivery) error {
	event, err := unmarshalEvent(msg.Body)
	if err != nil {
		return err
	}

	if err := c.saveStatistics(ctx, event); err != nil {
		return err
	}

	logger.Info("submission processed",
		zap.String("username", event.Username),
		zap.Time("timestamp", event.Timestamp.Truncate(time.Minute)),
	)

	return nil
}

func (c Consumer) saveStatistics(ctx context.Context, event events.Submission) error {
	truncatedTime := event.Timestamp.Truncate(time.Minute)

	if err := c.store.Increment(ctx, event.Username, truncatedTime); err != nil {
		return fmt.Errorf("failed to increment submission: %w", err)
	}

	return nil
}

func (c Consumer) Stop() error {
	close(c.stopChan)

	select {
	case <-c.doneChan:
	case <-time.After(2 * time.Second):
		logger.Warn("consumer stop timeout - consumer may not have been started")
	}

	closer.Close(c.channel, "failed to close consumer channel")
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func unmarshalEvent(body []byte) (events.Submission, error) {
	var event events.Submission
	if err := json.Unmarshal(body, &event); err != nil {
		return events.Submission{}, fmt.Errorf("failed to unmarshal event: %w", err)
	}
	return event, nil
}

func nackMessage(msg amqp.Delivery) {
	if err := msg.Nack(false, true); err != nil {
		logger.Error("failed to nack message", zap.Error(err))
	}
}

func ackMessage(msg amqp.Delivery) {
	if err := msg.Ack(false); err != nil {
		logger.Error("failed to ack message", zap.Error(err))
	}
}
