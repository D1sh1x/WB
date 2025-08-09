package handler

import (
	storage "WB2/internal/storage/postgres"
	"log/slog"
)

type Handler struct {
	log     *slog.Logger
	storage *storage.Storage
}

func NewHandler(log *slog.Logger, storage *storage.Storage) *Handler {
	return &Handler{
		log:     log,
		storage: storage,
	}
}
