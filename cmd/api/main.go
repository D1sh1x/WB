package main

import (
	"WB2/internal/config"
	"WB2/internal/lib/logger"
	"WB2/internal/server"
	storage "WB2/internal/storage/postgres"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.NewConfig()

	log := logger.SetupLogger(cfg.Env)

	log.Info("WB", slog.String("env", cfg.Env))
	log.Debug("debug message are enabled")

	db, err := storage.NewStorage(cfg.Database.DB_CONNECTION_STRING)
	if err != nil {
		log.Error("Failed to init storage", logger.Err(err))
		os.Exit(1)
	}

	server := server.NewServer(cfg)

	log.Info("Starting HTTP server", slog.String("port", cfg.HTTPServer.Port))

	go func() {
		if err := server.Start(log, db); err != nil && err != http.ErrServerClosed {
			log.Error("Failed to start server", logger.Err(err))
			os.Exit(1)
		}
	}()

	log.Info("Server started successfully", slog.String("port", cfg.HTTPServer.Port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		log.Error("Failed to stop server", logger.Err(err))
	}

	log.Info("Server gracefully stopped")

}
