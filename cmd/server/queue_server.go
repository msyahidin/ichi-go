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

// StartQueueWorkers starts all queue consumers for the configured driver.
// Blocks until ctx is cancelled.
func StartQueueWorkers(ctx context.Context, queueConfig *queue.Config, injector do.Injector) {
	if !queueConfig.Enabled {
		logger.Warnf("Queue system disabled — skipping worker startup")
		return
	}

	switch queueConfig.Driver {
	case "river":
		startRiverWorkers(ctx, injector)
	default:
		startRabbitMQWorkers(ctx, &queueConfig.RabbitMQ, injector)
	}
}

func startRiverWorkers(ctx context.Context, injector do.Injector) {
	client, err := do.Invoke[*riverqueue.Client[*sql.Tx]](injector)
	if err != nil || client == nil {
		logger.Errorf("River client unavailable — cannot start queue workers: %v", err)
		return
	}

	logger.Infof("🚀 Starting River queue workers...")
	if err := client.Start(ctx); err != nil {
		logger.Errorf("River client start error: %v", err)
		return
	}

	<-ctx.Done()

	logger.Infof("🛑 Stopping River queue workers...")
	if err := client.Stop(context.Background()); err != nil {
		logger.Errorf("River client stop error: %v", err)
	}
	logger.Infof("👋 River workers stopped")
}

func startRabbitMQWorkers(ctx context.Context, rabbitCfg *rabbitmq.Config, injector do.Injector) {
	conn, err := do.Invoke[*rabbitmq.Connection](injector)
	if conn == nil || err != nil {
		logger.Warnf("RabbitMQ connection unavailable — skipping worker startup")
		return
	}

	logger.Infof("🚀 Starting RabbitMQ queue workers...")

	// Declare all exchanges, queues, and bindings with exponential backoff retry.
	{
		backoff := 100 * time.Millisecond
		const maxBackoff = 10 * time.Second
		for {
			if err := rabbitmq.SetupTopology(conn, *rabbitCfg); err == nil {
				break
			} else {
				logger.Errorf("❌ Topology setup failed (retrying in %v): %v", backoff, err)
			}
			select {
			case <-ctx.Done():
				logger.Warnf("🛑 Context cancelled during topology setup — aborting")
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

	logger.Infof("✅ All RabbitMQ workers started")
	<-ctx.Done()
	logger.Infof("🛑 Shutting down RabbitMQ workers...")
	wg.Wait()
	logger.Infof("👋 All RabbitMQ workers stopped")
}
