package publisher

import (
	"context"
	"encoding/json"
	"fmt"

	"tcp-message-processor/internal/config"
	"tcp-message-processor/internal/server"
	"tcp-message-processor/pkg/broker"
	"tcp-message-processor/pkg/closer"
	"tcp-message-processor/pkg/logger"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Publisher struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
}

func New(cfg config.RabbitMQConfig) (*Publisher, error) {
	setup, err := broker.Publisher(broker.PublisherConfig{
		URL:      cfg.URL,
		Exchange: cfg.Exchange,
	})
	if err != nil {
		return nil, err
	}

	logger.Info("message broker publisher initialized", zap.String("exchange", cfg.Exchange))

	return &Publisher{
		conn:     setup.Conn,
		channel:  setup.Channel,
		exchange: cfg.Exchange,
	}, nil
}

func (p Publisher) Publish(ctx context.Context, event server.Event) error {
	body, err := marshal(event)
	if err != nil {
		return err
	}

	return p.publishMessage(ctx, body)
}

func (p Publisher) publishMessage(ctx context.Context, body []byte) error {
	err := p.channel.PublishWithContext(
		ctx,
		p.exchange,
		"submission.created",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

func (p Publisher) Close() error {
	closer.Close(p.channel, "failed to close publisher channel")
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

func marshal(event server.Event) ([]byte, error) {
	body, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}
	return body, nil
}
