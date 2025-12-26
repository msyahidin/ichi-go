package server

import (
	"context"
	queue "ichi-go/internal/infra/queue"
	"ichi-go/internal/infra/queue/rabbitmq"
	"ichi-go/pkg/logger"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// StartQueueWorkers starts all queue consumers.
// Renamed from "StartConsumer".
//
// Lifecycle:
// 1. Get registered consumers
// 2. Filter enabled consumers
// 3. Start worker pools
// 4. Wait for shutdown
// 5. Graceful stop
//
// Blocks until SIGTERM/SIGINT.
func StartQueueWorkers(queueConfig *queue.Config, conn *rabbitmq.Connection) {
	logger.Infof("🚀 Starting queue workers...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	wg := sync.WaitGroup{}

	// Get consumers
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

	<-sigChan
	logger.Infof("🛑 Shutting down...")

	cancel()

	logger.Infof("⏳ Waiting for workers...")
	wg.Wait()

	logger.Infof("👋 All workers stopped")
}
