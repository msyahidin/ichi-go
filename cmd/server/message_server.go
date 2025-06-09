package server

import (
	"github.com/labstack/echo/v4"
	"ichi-go/config"
	messageConfig "ichi-go/config/message"
	message "ichi-go/internal/infra/message/rabbitmq"
	"ichi-go/pkg/logger"
)

func SetupMessages(e *echo.Echo, messageConfig *messageConfig.MessageConfig) *message.ConnectionWrapper {
	// Initialize message broker connection
	if messageConfig.Enabled {
		logger.Debugf("initialized message rabbitmq configuration = %v", messageConfig.RabbitMQ)
		messageCfg := message.Config{
			URI: config.Message().RabbitMQ.GetRabbitMQURI(),
		}
		conn, err := message.New(messageCfg, logger.Log)
		if err != nil {
			logger.Fatalf("Failed to initialize RabbitMQ connection: %v", err)
		}
		logger.Debugf("Initialized RabbitMQ connection successfully")
		return conn
	} else {
		logger.Warnf("RabbitMQ is not enabled in the configuration")
	}
	return nil
}
