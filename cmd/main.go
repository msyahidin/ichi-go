package main

import (
	"context"
	"fmt"
	"github.com/uptrace/bun"
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
	mainConfig := config.LoadConfig(e)
	e.Debug = config.App().Debug
	middlewares.Init(e, mainConfig)
	logger.Init(e.Debug)

	//dbConnection := database.NewEntClient()
	dbConnection := database.NewBunClient(&mainConfig.Database)
	logger.Debugf("initialized database configuration = %v", dbConnection)

	//from docs define close on this function, but will impact cant create DB session on repository:
	defer func(dbConnection *bun.DB) {
		err := dbConnection.Close()
		if err != nil {
			logger.Fatalf("error initialized database configuration = %v", err)
		}
	}(dbConnection)

	cacheConnection := cache.New(&mainConfig.Cache)

	server.SetupRestRoutes(e, mainConfig, dbConnection, cacheConnection)
	server.SetupWebRoutes(e, mainConfig)
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
		address := fmt.Sprintf(":%d", config.Http().Port)
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
