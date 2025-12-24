package main

import (
	"context"
	"fmt"
	"ichi-go/cmd/server"
	"ichi-go/config"
	"ichi-go/internal/infra"
	"ichi-go/internal/infra/messaging/rabbitmq"
	"ichi-go/internal/middlewares"
	"ichi-go/pkg/errorhandler"
	"ichi-go/pkg/logger"
	"os"
	"os/signal"
	"time"

	"github.com/samber/do/v2"

	"github.com/labstack/echo/v4"
)

func main() {
	injector := do.New()

	e := echo.New()
	cfg := config.MustLoad()
	logger.Debugf("initialized configuration %+v", *cfg)
	if cfg == nil {
		logger.Fatalf("failed to load configuration")
	}
	config.SetDebugMode(e, cfg.App().Debug)
	infra.Setup(injector, cfg)
	middlewares.Init(e, cfg)
	logger.GetInstance()

	// Setup messaging if enabled
	if cfg.Messaging().Enabled {
		msgConfig := cfg.Messaging()
		msgConn := do.MustInvoke[*rabbitmq.Connection](injector)
		server.StartConsumer(msgConfig, msgConn)
	}

	// Setup web routes and error handler
	server.SetupRestRoutes(injector, e, cfg)
	server.SetupWebRoutes(e, cfg.Schema())
	errorhandler.Setup(e)

	// Log all routes
	for _, route := range e.Routes() {
		if route.Method == "" && route.Path == "" {
			continue
		}
		logger.Debugf("Routes Mapped: %s %s", route.Method, route.Path)
	}

	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Start the server
	go func() {
		address := fmt.Sprintf(":%d", cfg.Http().Port)
		logger.Infof("starting http server at %s", address)
		if err := e.Start(address); err != nil {
			logger.Fatalf("shutting down the rest server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		logger.Fatalf("error during server shutdown: %v", err)
	}

	// Shutdown all services in reverse dependency order
	logger.Infof("shutting down services...")
	injector.Shutdown()
	logger.Infof("goodbye!")
}

// Deprecated
//func mainLegacy() {
//
//	e := echo.New()
//	cfg := config.MustLoad()
//	if cfg == nil {
//		logger.Fatalf("failed to load configuration")
//	}
//	config.SetDebugMode(e, cfg.App().Debug)
//	middlewares.Init(e, cfg)
//	logger.GetInstance()
//	//dbConnection := database.NewEntClient()
//	dbConnection, _ := database.NewBunClient(cfg.Database())
//	logger.Debugf("initialized database configuration = %v", dbConnection)
//
//	//from docs define close on this function, but will impact cant create DB session on repository:
//	defer func(dbConnection *bun.DB) {
//		err := dbConnection.Close()
//		if err != nil {
//			logger.Fatalf("error initialized database configuration = %v", err)
//		}
//	}(dbConnection)
//
//	//cacheConnection := cache.New(cfg.Cache())
//
//	msgConnection, err := rabbitmq.NewConnection(cfg.Messaging().RabbitMQ)
//
//	if cfg.Messaging().Enabled {
//		if err != nil {
//			logger.Fatalf("Failed to connect: %+v", err)
//		}
//		defer msgConnection.Close()
//
//		server.StartConsumer(cfg.Messaging(), msgConnection)
//	}
//
//	//server.SetupRestRoutes(e, cfg, dbConnection, cacheConnection, msgConnection)
//	server.SetupWebRoutes(e, cfg.Schema())
//	errorhandler.Setup(e)
//
//	for _, route := range e.Routes() {
//		if route.Method == "" && route.Path == "" {
//			continue
//		}
//		logger.Debugf("Routes Mapped: %s %s", route.Method, route.Path)
//	}
//
//	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
//	defer stop()
//
//	// Start the server
//	go func() {
//		address := fmt.Sprintf(":%d", cfg.Http().Port)
//		logger.Infof("starting http server at %s", address)
//		if err := e.Start(address); err != nil {
//			logger.Fatalf("shutting down the rest server: %v", err)
//		}
//	}()
//
//	// Wait for interrupt signal to gracefully shut down the server with a timeout of 10 seconds.
//	<-ctx.Done()
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//	if err := e.Shutdown(ctx); err != nil {
//		logger.Fatalf("shutting down the rest server: %v", err)
//	}
//}
