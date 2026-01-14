package broker

import (
	"fmt"

	"tcp-message-processor/common/pkg/closer"

	amqp "github.com/rabbitmq/amqp091-go"
)

type (
	ConsumerConfig struct {
		URL        string
		Exchange   string
		Queue      string
		RoutingKey string
	}

	PublisherConfig struct {
		URL      string
		Exchange string
	}

	Setup struct {
		Conn      *amqp.Connection
		Channel   *amqp.Channel
		QueueName string
	}
)

func connect(url string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to message broker: %w", err)
	}
	return conn, nil
}

func openChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}
	return channel, nil
}

func configExchange(channel *amqp.Channel, exchange string) error {
	err := channel.ExchangeDeclare(
		exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}
	return nil
}

func configQueue(channel *amqp.Channel, queueName string) (amqp.Queue, error) {
	queue, err := channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("failed to declare queue: %w", err)
	}
	return queue, nil
}

func bindQueue(channel *amqp.Channel, queueName, exchange, routingKey string) error {
	err := channel.QueueBind(
		queueName,
		routingKey,
		exchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}
	return nil
}

func setQoS(channel *amqp.Channel, prefetchCount int) error {
	err := channel.Qos(prefetchCount, 0, false)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}
	return nil
}

func Consumer(cfg ConsumerConfig) (*Setup, error) {
	conn, err := connect(cfg.URL)
	if err != nil {
		return nil, err
	}

	channel, err := openChannel(conn)
	if err != nil {
		closer.Close(conn, "failed to close connection during consumer setup")
		return nil, err
	}

	if err := configExchange(channel, cfg.Exchange); err != nil {
		closer.Close(channel, "failed to close channel during consumer setup")
		closer.Close(conn, "failed to close connection during consumer setup")
		return nil, err
	}

	if err := setQoS(channel, 1); err != nil {
		closer.Close(channel, "failed to close channel during consumer setup")
		closer.Close(conn, "failed to close connection during consumer setup")
		return nil, err
	}

	queue, err := configQueue(channel, cfg.Queue)
	if err != nil {
		closer.Close(channel, "failed to close channel during consumer setup")
		closer.Close(conn, "failed to close connection during consumer setup")
		return nil, err
	}

	if err := bindQueue(channel, queue.Name, cfg.Exchange, cfg.RoutingKey); err != nil {
		closer.Close(channel, "failed to close channel during consumer setup")
		closer.Close(conn, "failed to close connection during consumer setup")
		return nil, err
	}

	return &Setup{
		Conn:      conn,
		Channel:   channel,
		QueueName: queue.Name,
	}, nil
}

func Publisher(cfg PublisherConfig) (*Setup, error) {
	conn, err := connect(cfg.URL)
	if err != nil {
		return nil, err
	}

	channel, err := openChannel(conn)
	if err != nil {
		closer.Close(conn, "failed to close connection during publisher setup")
		return nil, err
	}

	if err := configExchange(channel, cfg.Exchange); err != nil {
		closer.Close(channel, "failed to close channel during publisher setup")
		closer.Close(conn, "failed to close connection during publisher setup")
		return nil, err
	}

	return &Setup{
		Conn:    conn,
		Channel: channel,
	}, nil
}
