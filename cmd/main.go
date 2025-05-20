package main

import (
	"context"
	"fmt"
	"ichi-go/cmd/server"
	"ichi-go/config"
	"ichi-go/internal/infra/cache"
	"ichi-go/internal/infra/database"
	"ichi-go/internal/infra/database/ent"
	"ichi-go/internal/middlewares"
	errorhandler "ichi-go/pkg/errorhandler"
	"ichi-go/pkg/logger"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
)

func main() {
	config.LoadConfig()
	logger.Init()

	e := echo.New()
	middlewares.Init(e)

	dbConnection := database.NewEntClient()
	logger.Debugf("initialized database configuration = %v", dbConnection)

	//from docs define close on this function, but will impact cant create DB session on repository:
	defer func(dbConnection *ent.Client) {
		err := dbConnection.Close()
		if err != nil {
			logger.Fatalf("error initialized database configuration = %v", err)
		}
	}(dbConnection)

	cacheConnection := cache.New()

	server.SetupRestRoutes(e, dbConnection, cacheConnection)
	server.SetupWebRoutes(e)
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
			logger.Fatalf("shutting down the rest server")
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
