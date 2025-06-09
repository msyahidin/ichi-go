package rabbitmq

import (
	"context"
	"ichi-go/pkg/logger"
	"sync"
	"sync/atomic"

	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
)

type ConnectionWrapper struct {
	cfg    Config
	logger logger.Logger

	conn   *amqp.Connection
	lock   sync.Mutex
	close  chan struct{}
	once   uint32
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func New(cfg Config, logger logger.Logger) (*ConnectionWrapper, error) {

	w := &ConnectionWrapper{
		cfg:    cfg,
		logger: logger,
		close:  make(chan struct{}),
	}

	if err := w.connect(); err != nil {
		return nil, err
	}

	go w.handleReconnect()

	return w, nil
}

func (w *ConnectionWrapper) connect() error {
	w.lock.Lock()
	defer w.lock.Unlock()

	var (
		conn *amqp.Connection
		err  error
	)

	if w.cfg.TLSConfig != nil {
		conn, err = amqp.DialTLS(w.cfg.URI, w.cfg.TLSConfig)
	} else {
		conn, err = amqp.Dial(w.cfg.URI)
	}
	if err != nil {
		return errors.Wrap(err, "failed to connect to RabbitMQ")
	}

	w.conn = conn
	w.logger.Debugf("Connected to RabbitMQ")

	return nil
}

func (w *ConnectionWrapper) Connection() *amqp.Connection {
	return w.conn
}

func (w *ConnectionWrapper) IsConnected() bool {
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.conn != nil && !w.conn.IsClosed()
}

func (w *ConnectionWrapper) Close() error {
	if !atomic.CompareAndSwapUint32(&w.once, 0, 1) {
		return nil
	}
	close(w.close)
	w.wg.Wait()
	return w.conn.Close()
}

func (w *ConnectionWrapper) handleReconnect() {
	w.wg.Add(1)
	defer w.wg.Done()

	errChan := w.conn.NotifyClose(make(chan *amqp.Error))

	for {
		select {
		case <-w.close:
			return
		case err := <-errChan:
			if err != nil {
				w.logger.Error("RabbitMQ connection closed, reconnecting...", err, nil)

				b := w.cfg.ReconnectBackoff
				retryError := backoff.Retry(func() error {
					return w.connect()
				}, b)

				if retryError != nil {
					w.logger.Error("Failed to reconnect after retries", err, nil)
					// Consider circuit breaker pattern here
				}

				errChan = w.conn.NotifyClose(make(chan *amqp.Error))
			}
		}
	}

}
