package handler

import (
	"log/slog"

	"WB2/internal/cache"
	storage "WB2/internal/storage/postgres"
)

type Handler struct {
	log     *slog.Logger
	storage *storage.Storage
	cache   *cache.OrderCache
}

func NewHandler(log *slog.Logger, storage *storage.Storage, cache *cache.OrderCache) *Handler {
	return &Handler{
		log:     log,
		storage: storage,
		cache:   cache,
	}
}
