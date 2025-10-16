package rabbitmq

import (
	"fmt"
	messagingConfig "ichi-go/config/messaging"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Connection struct {
	config messagingConfig.RabbitConnectionConfig
	conn   *amqp.Connection
	mu     sync.RWMutex
	closed bool
}

func NewConnection(config messagingConfig.MessagingConfig) (*Connection, error) {
	c := &Connection{
		config: config.RabbitMQ.Connection,
		closed: false,
	}

	if err := c.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	go c.handleReconnect()

	log.Printf("✅ RabbitMQ connected: %s:%d", c.config.Host, c.config.Port)
	return c, nil
}

func (c *Connection) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, err := amqp.Dial(messagingConfig.GetRabbitMQURI(c.config))
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

		log.Printf("⚠️ Connection lost: %v. Reconnecting...", reason)

		for attempt := 0; attempt < 10; attempt++ {
			if c.closed {
				return
			}

			time.Sleep(time.Duration(attempt*2) * time.Second)

			if err := c.connect(); err != nil {
				log.Printf("Reconnect attempt %d failed: %v", attempt+1, err)
				continue
			}

			log.Println("✅ Reconnected successfully")
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
