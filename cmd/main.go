package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/samber/do/v2"

	"ichi-go/cmd/server"
	"ichi-go/config"
	_ "ichi-go/docs" // Import generated docs
	"ichi-go/internal/infra"
	"ichi-go/internal/infra/queue/rabbitmq"
	"ichi-go/internal/middlewares"
	errors2 "ichi-go/pkg/errors"
	"ichi-go/pkg/logger"
)

//	@title			Ichi-Go API
//	@version		1.0
//	@description	A production-ready Go backend template built on clean architecture principles.
//	@description	Features include JWT authentication with multiple algorithms (HMAC, RSA, ECDSA),
//	@description	multi-language validation (English/Indonesian), RabbitMQ message queuing,
//	@description	Redis caching with LZ4 compression, and comprehensive error handling.

//	@contact.name	API Support
//	@contact.email	support@ichi-go.dev

//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT

//	@host		localhost:8080
//	@BasePath	/ichi-go/api

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

//	@tag.name			Auth
//	@tag.description	Authentication and authorization endpoints including login, register, token refresh, and user profile

//	@tag.name			Users
//	@tag.description	User management endpoints

//	@externalDocs.description	OpenAPI Specification
//	@externalDocs.url			https://swagger.io/specification/

func main() {
	injector := do.New()

	e := echo.New()
	cfg := config.MustLoad()
	logger.Debugf("initialized configuration %+v", *cfg)
	if cfg == nil || cfg.Schema() == nil {
		logger.Fatalf("failed to load configuration")
	}
	config.SetDebugMode(e, cfg.App().Debug)
	infra.Setup(injector, cfg)
	middlewares.Init(e, cfg)
	logger.GetInstance()

	// Setup web routes and error handler
	server.SetupRestRoutes(injector, e, cfg)
	server.SetupWebRoutes(e, cfg.Schema())
	errors2.Setup(e)

	// Log all routes
	for _, route := range e.Router().Routes() {
		if route.Method == "" && route.Path == "" {
			continue
		}
		logger.Debugf("Routes Mapped: %s %s", route.Method, route.Path)
	}

	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Setup messaging if enabled
	if cfg.Queue().Enabled {
		msgConfig := cfg.Queue()
		msgConn := do.MustInvoke[*rabbitmq.Connection](injector)
		go server.StartQueueWorkers(ctx, msgConfig, msgConn, injector)
	}

	// Start the server with context-based graceful shutdown
	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		address := fmt.Sprintf(":%d", cfg.Http().Port)
		logger.Infof("starting http server at %s", address)
		sc := echo.StartConfig{
			Address:         address,
			GracefulTimeout: 10 * time.Second,
			OnShutdownError: func(err error) {
				logger.Errorf("error during server shutdown: %v", err)
			},
		}
		if err := sc.Start(ctx, e); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("http server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()

	// Graceful shutdown
	logger.Infof("Received shutdown signal...")

	// Wait for server graceful shutdown to complete
	<-serverDone

	// Shutdown all services in reverse dependency order
	logger.Infof("shutting down services...")
	injector.Shutdown()
	logger.Infof("goodbye!")
}
