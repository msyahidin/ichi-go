package server

import (
	"context"
	queue "ichi-go/internal/infra/queue"
	"ichi-go/internal/infra/queue/rabbitmq"
	"ichi-go/pkg/logger"
	"sync"
)

// StartQueueWorkers starts all queue consumers with context-based lifecycle.
// Blocks until context is cancelled.
func StartQueueWorkers(ctx context.Context, queueConfig *queue.Config, conn *rabbitmq.Connection) {
	if conn == nil {
		logger.Warnf("Queue connection is nil - skipping worker startup")
		return
	}

	logger.Infof("🚀 Starting queue workers...")

	wg := sync.WaitGroup{}

	// Get registered consumers
	registeredConsumers := queue.GetRegisteredConsumers()

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
		go func(name string, consumer rabbitmq.MessageConsumer, consumeFunc rabbitmq.ConsumeFunc, desc string) {
			defer wg.Done()

			logger.Infof("✅ Started %s: %s", name, desc)

			if err := consumer.Consume(ctx, consumeFunc); err != nil {
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
