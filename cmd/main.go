package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"rathalos-kit/cmd/server"
	"rathalos-kit/config"
	"rathalos-kit/internal/infrastructure/database"
	"rathalos-kit/internal/infrastructure/database/ent"
	"rathalos-kit/internal/middlewares"
	"rathalos-kit/pkg/logger"
	"time"

	"github.com/labstack/echo/v4"
)

func main() {
	config.LoadConfig()
	logger.Init()

	e := echo.New()
	middlewares.Init(e)

	dbConnection := database.NewEntClient()
	logger.Log.Debug().Any("initialized database configuration=", dbConnection)

	//from docs define close on this function, but will impact cant create DB session on repository:
	defer func(dbConnection *ent.Client) {
		err := dbConnection.Close()
		if err != nil {
			logger.Log.Fatal().Err(err).Msg("error initialized database configuration")
		}
	}(dbConnection)

	server.SetupRestRoutes(e, dbConnection)
	server.SetupWebRoutes(e)

	for _, route := range e.Routes() {
		logger.Log.Debug().Msgf("Routes Mapped: %s %s", route.Method, route.Path)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Start the server
	go func() {
		address := fmt.Sprintf(":%d", config.Cfg.Http.Port)
		e.Logger.Info("REST API server running on", address)
		if err := e.Start(address); err != nil {
			e.Logger.Fatal("shutting down the rest server")
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server with a timeout of 10 seconds.
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
