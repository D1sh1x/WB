package main

import (
	"WB2/internal/cache"
	"WB2/internal/config"
	"WB2/internal/kafka"
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

	// init cache and warm-up from DB
	orderCache := cache.NewOrderCache()
	if orders, err := db.GetAllOrders(); err == nil {
		orderCache.Load(orders)
		log.Info("cache warmed", slog.Int("orders", len(orders)))
	} else {
		log.Warn("failed to warm cache", logger.Err(err))
	}

	srv := server.NewServer(cfg, orderCache)

	log.Info("Starting HTTP server", slog.String("port", cfg.HTTPServer.Port))

	// start HTTP server
	go func() {
		if err := srv.Start(log, db); err != nil && err != http.ErrServerClosed {
			log.Error("Failed to start server", logger.Err(err))
			os.Exit(1)
		}
	}()

	// start Kafka consumer
	var consumerCancel context.CancelFunc
	if len(cfg.Kafka.Brokers) > 0 && cfg.Kafka.Topic != "" && cfg.Kafka.GroupID != "" {
		ctx, cancel := context.WithCancel(context.Background())
		consumerCancel = cancel
		cons, err := kafka.NewConsumer(log, db, orderCache, cfg.Kafka.Brokers, cfg.Kafka.GroupID, cfg.Kafka.Topic, cfg.Kafka.Version)
		if err != nil {
			log.Error("Failed to init kafka consumer", logger.Err(err))
		} else {
			go func() {
				if err := cons.Run(ctx); err != nil && err != context.Canceled {
					log.Error("Kafka consumer stopped", logger.Err(err))
				}
			}()
		}
	}

	log.Info("Server started successfully", slog.String("port", cfg.HTTPServer.Port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if consumerCancel != nil {
		consumerCancel()
	}

	if err := srv.Stop(ctx); err != nil {
		log.Error("Failed to stop server", logger.Err(err))
	}

	log.Info("Server gracefully stopped")

}
