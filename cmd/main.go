package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"rathalos-kit/cmd/server"
	"rathalos-kit/config"
	"rathalos-kit/internal/logger"
	"rathalos-kit/internal/middlewares"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
)

func main() {
	config.LoadConfig()
	logger.Init()

	e := echo.New()
	middlewares.Init(e)

	server.SetupRestRoutes(e)
	server.SetupWebRoutes(e)

	go func() {
		address := fmt.Sprintf(":%d", config.AppConfig.Server.Rest.Port)
		e.Logger.Info("REST API server running on", address)
		if err := e.Start(address); err != nil {
			e.Logger.Fatal("shutting down the rest server")
		}
	}()
	//
	//go func() {
	//	address := fmt.Sprintf(":%d", config.AppConfig.Server.Web.Port)
	//	e.Logger.Info("Web Server running on", address)
	//	if err := e.Start(address); err != nil {
	//		e.Logger.Fatal("shutting down the web server")
	//	}
	//}()

	// Wait for interrupt signal to gracefully shut down the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
