package main

import (
	"context"
	"fmt"
	"github.com/uptrace/bun"
	"ichi-go/cmd/consumer"
	"ichi-go/cmd/server"
	"ichi-go/config"
	"ichi-go/internal/infra/cache"
	"ichi-go/internal/infra/database"
	"ichi-go/internal/middlewares"
	"ichi-go/pkg/errorhandler"
	"ichi-go/pkg/logger"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
)

func main() {

	e := echo.New()
	config.MustLoad()
	cfg := config.Get()
	if cfg == nil {
		logger.Fatalf("failed to load configuration")
	}
	config.SetDebugMode(e, cfg.App().Debug)
	middlewares.Init(e, cfg)
	logger.Init(e.Debug, cfg.App().Env == "local" || cfg.App().Env == "development")
	//dbConnection := database.NewEntClient()
	dbConnection, _ := database.NewBunClient(cfg.Database())
	logger.Debugf("initialized database configuration = %v", dbConnection)

	//from docs define close on this function, but will impact cant create DB session on repository:
	defer func(dbConnection *bun.DB) {
		err := dbConnection.Close()
		if err != nil {
			logger.Fatalf("error initialized database configuration = %v", err)
		}
	}(dbConnection)

	cacheConnection := cache.New(cfg.Cache())

	server.SetupRestRoutes(e, cfg, dbConnection, cacheConnection)
	server.SetupWebRoutes(e, cfg.Schema())
	if cfg.Messaging().Enabled {
		consumer.Start(cfg.Messaging())
	}
	errorhandler.Setup(e)

	for _, route := range e.Routes() {
		if route.Method == "" && route.Path == "" {
			continue
		}
		logger.Debugf("Routes Mapped: %s %s", route.Method, route.Path)
	}

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

	// Wait for interrupt signal to gracefully shut down the server with a timeout of 10 seconds.
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		logger.Fatalf("shutting down the rest server: %v", err)
	}
}
