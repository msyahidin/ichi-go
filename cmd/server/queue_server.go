package server

import (
	"context"
	"sync"
	"time"

	"github.com/samber/do/v2"

	queue "ichi-go/internal/infra/queue"
	"ichi-go/internal/infra/queue/rabbitmq"
	"ichi-go/pkg/logger"
)

// StartQueueWorkers starts all queue consumers with context-based lifecycle.
// Blocks until context is cancelled.
func StartQueueWorkers(ctx context.Context, queueConfig *queue.Config, conn *rabbitmq.Connection, injector do.Injector) {
	if conn == nil {
		logger.Warnf("Queue connection is nil - skipping worker startup")
		return
	}

	logger.Infof("🚀 Starting queue workers...")

	// Declare all exchanges, queues, and bindings once before any producer or consumer starts.
	// Retry with exponential backoff so transient broker-not-ready errors at startup self-heal.
	{
		backoff := 100 * time.Millisecond
		const maxBackoff = 10 * time.Second
		for {
			if err := rabbitmq.SetupTopology(conn, queueConfig.RabbitMQ); err == nil {
				break
			} else {
				logger.Errorf("❌ Topology setup failed (retrying in %v): %v", backoff, err)
			}
			select {
			case <-ctx.Done():
				logger.Warnf("🛑 Context cancelled during topology setup — aborting worker startup")
				return
			case <-time.After(backoff):
				backoff = min(backoff*2, maxBackoff)
			}
		}
	}

	wg := sync.WaitGroup{}

	// Get registered consumers — pass injector so consumers can resolve DB/service deps.
	registeredConsumers := queue.GetRegisteredConsumers(injector)

	// Start each consumer
	for _, registration := range registeredConsumers {
		consumerCfg, err := rabbitmq.GetConsumerByName(&queueConfig.RabbitMQ, registration.Name)
		if err != nil {
			logger.Infof("⏭️  Skipping %s: %v", registration.Name, err)
			continue
		}

		if !consumerCfg.Enabled {
			logger.Infof("⏭️  Disabled: %s", registration.Name)
			continue
		}

		exchangeCfg, err := rabbitmq.GetExchangeByName(&queueConfig.RabbitMQ, consumerCfg.ExchangeName)
		if err != nil {
			logger.Errorf("❌ No exchange for %s: %v", registration.Name, err)
			continue
		}

		consumer, err := rabbitmq.NewConsumer(conn, *consumerCfg, *exchangeCfg)
		if err != nil {
			logger.Errorf("❌ Failed to create %s: %v", registration.Name, err)
			continue
		}

		wg.Add(1)
		go func(name string, consumer rabbitmq.MessageConsumer, consumeFunc queue.ConsumeFunc, desc string) {
			defer wg.Done()

			logger.Infof("✅ Started %s: %s", name, desc)

			if err := consumer.Consume(ctx, rabbitmq.ConsumeFunc(consumeFunc)); err != nil {
				logger.Errorf("❌ %s error: %v", name, err)
			}

			logger.Infof("👋 Stopped %s", name)
		}(registration.Name, consumer, registration.ConsumeFunc, registration.Description)
	}

	logger.Infof("✅ All workers started")

	// Wait for context cancellation
	<-ctx.Done()

	logger.Infof("🛑 Shutting down queue workers...")

	// Wait for all workers to finish
	logger.Infof("⏳ Waiting for workers to finish...")
	wg.Wait()

	logger.Infof("👋 All queue workers stopped")
}
