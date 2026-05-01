package server

import (
	"context"
	"database/sql"
	"sync"
	"time"

	riverqueue "github.com/riverqueue/river"
	"github.com/samber/do/v2"

	queue "ichi-go/internal/infra/queue"
	"ichi-go/internal/infra/queue/rabbitmq"
	"ichi-go/pkg/logger"
)

// StartQueueWorkers starts workers for every enabled queue connection concurrently.
// Blocks until ctx is cancelled and all workers have shut down.
func StartQueueWorkers(ctx context.Context, queueCfg *queue.QueueSchema, injector do.Injector) {
	enabled := queueCfg.EnabledConnections()
	if len(enabled) == 0 {
		logger.Warnf("Queue system disabled — skipping worker startup")
		return
	}

	var wg sync.WaitGroup
	for _, nc := range enabled {
		nc := nc
		wg.Add(1)
		go func() {
			defer wg.Done()
			switch nc.Config.Driver {
			case "amqp":
				startAMQPWorkers(ctx, nc.Name, &nc.Config.AMQP, injector)
			case "database":
				startRiverWorkers(ctx, nc.Name, injector)
			default:
				logger.Errorf("unknown queue driver %q for connection %q", nc.Config.Driver, nc.Name)
			}
		}()
	}
	wg.Wait()
}

func startRiverWorkers(ctx context.Context, connName string, injector do.Injector) {
	client, err := do.InvokeNamed[*riverqueue.Client[*sql.Tx]](injector, "queue.river."+connName)
	if err != nil || client == nil {
		logger.Errorf("River client unavailable for %q — cannot start queue workers: %v", connName, err)
		return
	}

	logger.Infof("🚀 Starting River queue workers [%s]...", connName)
	if err := client.Start(ctx); err != nil {
		logger.Errorf("River client start error [%s]: %v", connName, err)
		return
	}

	<-ctx.Done()

	logger.Infof("🛑 Stopping River queue workers [%s]...", connName)
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer stopCancel()
	if err := client.Stop(stopCtx); err != nil {
		logger.Errorf("River client stop error [%s]: %v", connName, err)
	}
	logger.Infof("👋 River workers stopped [%s]", connName)
}

func startAMQPWorkers(ctx context.Context, connName string, rabbitCfg *rabbitmq.Config, injector do.Injector) {
	conn, err := do.InvokeNamed[*rabbitmq.Connection](injector, "queue.conn."+connName)
	if conn == nil || err != nil {
		logger.Warnf("RabbitMQ connection unavailable for %q — skipping worker startup", connName)
		return
	}

	logger.Infof("🚀 Starting RabbitMQ queue workers [%s]...", connName)

	// Declare all exchanges, queues, and bindings with exponential backoff retry.
	{
		backoff := 100 * time.Millisecond
		const maxBackoff = 10 * time.Second
		for {
			if err := rabbitmq.SetupTopology(conn, *rabbitCfg); err == nil {
				break
			} else {
				logger.Errorf("❌ Topology setup failed [%s] (retrying in %v): %v", connName, backoff, err)
			}
			select {
			case <-ctx.Done():
				logger.Warnf("🛑 Context cancelled during topology setup [%s] — aborting", connName)
				return
			case <-time.After(backoff):
				backoff = min(backoff*2, maxBackoff)
			}
		}
	}

	wg := sync.WaitGroup{}
	registeredConsumers := queue.GetRegisteredConsumers(injector)

	for _, registration := range registeredConsumers {
		consumerCfg, err := rabbitmq.GetConsumerByName(rabbitCfg, registration.Name)
		if err != nil {
			logger.Infof("⏭️  Skipping %s: %v", registration.Name, err)
			continue
		}
		if !consumerCfg.Enabled {
			logger.Infof("⏭️  Disabled: %s", registration.Name)
			continue
		}
		exchangeCfg, err := rabbitmq.GetExchangeByName(rabbitCfg, consumerCfg.ExchangeName)
		if err != nil {
			logger.Errorf("❌ No exchange for %s: %v", registration.Name, err)
			continue
		}
		consumer, err := rabbitmq.NewConsumer(conn, *consumerCfg, *exchangeCfg)
		if err != nil {
			logger.Errorf("❌ Failed to create consumer %s: %v", registration.Name, err)
			continue
		}

		wg.Add(1)
		go func(name string, c rabbitmq.MessageConsumer, fn queue.ConsumeFunc, desc string) {
			defer wg.Done()
			logger.Infof("✅ Started %s: %s", name, desc)
			if err := c.Consume(ctx, rabbitmq.ConsumeFunc(fn)); err != nil {
				logger.Errorf("❌ %s error: %v", name, err)
			}
			logger.Infof("👋 Stopped %s", name)
		}(registration.Name, consumer, registration.ConsumeFunc, registration.Description)
	}

	logger.Infof("✅ All RabbitMQ workers started [%s]", connName)
	<-ctx.Done()
	logger.Infof("🛑 Shutting down RabbitMQ workers [%s]...", connName)
	wg.Wait()
	logger.Infof("👋 All RabbitMQ workers stopped [%s]", connName)
}
