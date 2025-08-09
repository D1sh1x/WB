package handler

import (
	"log/slog"

	"WB2/internal/kafka"
	storage "WB2/internal/storage/postgres"
)

type Handler struct {
	log     *slog.Logger
	storage *storage.Storage
	kafka   *kafka.Service
}

func NewHandler(log *slog.Logger, storage *storage.Storage, kafka *kafka.Service) *Handler {
	return &Handler{
		log:     log,
		storage: storage,
		kafka:   kafka,
	}
}
