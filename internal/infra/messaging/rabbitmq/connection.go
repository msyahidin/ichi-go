package rabbitmq

import (
	"fmt"
	"ichi-go/pkg/logger"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Connection struct {
	config RabbitMQConfig
	conn   *amqp.Connection
	mu     sync.RWMutex
	closed bool
}

func NewConnection(config RabbitMQConfig) (*Connection, error) {
	c := &Connection{
		config: config,
		closed: false,
	}

	if err := c.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	go c.handleReconnect()

	logger.Infof("RabbitMQ connected: %s:%d", config.Connection.Host, config.Connection.Port)
	return c, nil
}

func (c *Connection) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, err := amqp.Dial(GetRabbitMQURI(c.config.Connection))
	if err != nil {
		return err
	}

	c.conn = conn
	return nil
}

func (c *Connection) handleReconnect() {
	for {
		if c.closed {
			return
		}

		reason, ok := <-c.conn.NotifyClose(make(chan *amqp.Error))
		if !ok || c.closed {
			return
		}

		logger.Infof("Connection lost: %v. Reconnecting...", reason)

		for attempt := 0; attempt < 10; attempt++ {
			if c.closed {
				return
			}

			time.Sleep(time.Duration(attempt*2) * time.Second)

			if err := c.connect(); err != nil {
				logger.Infof("Reconnect attempt %d failed: %v", attempt+1, err)
				continue
			}

			logger.Infof("Reconnected successfully")
			break
		}
	}
}

func (c *Connection) GetConnection() *amqp.Connection {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}

func (c *Connection) Close() error {
	c.mu.Lock()
	c.closed = true
	c.mu.Unlock()

	if c.conn != nil && !c.conn.IsClosed() {
		return c.conn.Close()
	}
	return nil
}
