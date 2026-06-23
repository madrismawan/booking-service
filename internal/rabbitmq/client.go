package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	mu   sync.Mutex
}

func NewClient(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	if err := ch.Confirm(false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	return &Client{conn: conn, ch: ch}, nil
}

func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) DeclareQueue(name string) (amqp.Queue, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.declareQueue(name)
}

func (c *Client) declareQueue(name string) (amqp.Queue, error) {
	return c.ch.QueueDeclare(
		name,
		true,
		false,
		false,
		false,
		nil,
	)
}

func (c *Client) PublishJSON(ctx context.Context, queueName string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, err := c.declareQueue(queueName); err != nil {
		return err
	}

	confirmation, err := c.ch.PublishWithDeferredConfirmWithContext(
		ctx,
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
	if err != nil {
		return err
	}

	acked, err := confirmation.WaitContext(ctx)
	if err != nil {
		return err
	}
	if !acked {
		return errors.New("rabbitmq broker rejected published message")
	}
	return nil
}

func (c *Client) Consume(queueName, consumerName string) (<-chan amqp.Delivery, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, err := c.declareQueue(queueName); err != nil {
		return nil, err
	}

	if err := c.ch.Qos(1, 0, false); err != nil {
		return nil, err
	}

	return c.ch.Consume(
		queueName,
		consumerName,
		false,
		false,
		false,
		false,
		nil,
	)
}
