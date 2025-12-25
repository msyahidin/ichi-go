package server

import (
	"context"
	infraMessaging "ichi-go/internal/infra/messaging"
	"ichi-go/internal/infra/messaging/rabbitmq"
	"ichi-go/pkg/logger"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func StartConsumer(msgConfig *infraMessaging.Config, conn *rabbitmq.Connection) {
	logger.Infof("Starting consumer workers...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	wg := sync.WaitGroup{}
	consumers := infraMessaging.GetRegisteredConsumers()

	for _, c := range consumers {
		consumerCfg, err := rabbitmq.GetConsumerByName(&msgConfig.RabbitMQ, c.Name)
		if err != nil {
			logger.Infof("Skipping %s: %v", c.Name, err)
			continue
		}

		if !consumerCfg.Enabled {
			logger.Infof("Consumer '%s' is disabled", c.Name)
			continue
		}

		exchangeCfg, err := rabbitmq.GetExchangeByName(&msgConfig.RabbitMQ, consumerCfg.ExchangeName)
		if err != nil {
			logger.Infof("Skipping %s: %v", c.Name, err)
			continue
		}

		consumer, err := rabbitmq.NewConsumer(conn, *consumerCfg, *exchangeCfg)
		if err != nil {
			logger.Errorf("Failed to create consumer %s: %v", c.Name, err)
			continue
		}

		wg.Add(1)
		go func(name string, consumer rabbitmq.MessageConsumer, handler rabbitmq.MessageHandler) {
			defer wg.Done()

			if err := consumer.Consume(ctx, handler); err != nil {
				logger.Infof("Consumer '%s' error: %v", name, err)
			}
		}(c.Name, consumer, c.Handler)
	}

	logger.Infof("All consumers started")

	<-sigChan
	logger.Infof("Shutting down gracefully...")

	cancel()
	wg.Wait()

	logger.Infof("Goodbye!")
}
