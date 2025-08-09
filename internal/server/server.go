package server

import (
	"context"
	"log/slog"
	"net/http"

	"WB2/internal/cache"
	"WB2/internal/config"
	storage "WB2/internal/storage/postgres"

	"github.com/labstack/echo/v4"
)

// Server инкапсулирует echo-сервер и конфигурацию
type Server struct {
	cfg    *config.Config
	router *echo.Echo
	server *http.Server
	cache  *cache.OrderCache
}

func NewServer(cfg *config.Config, c *cache.OrderCache) *Server {
	return &Server{
		cfg:    cfg,
		router: echo.New(),
		cache:  c,
	}
}

// Start запускает HTTP-сервер и регистрирует маршруты
func (s *Server) Start(log *slog.Logger, storage *storage.Storage) error {
	InitRoutes(s.router, log, storage, s.cfg, s.cache)
	s.server = &http.Server{
		Addr:         ":" + s.cfg.HTTPServer.Port,
		Handler:      s.router,
		ReadTimeout:  s.cfg.HTTPServer.Timeout,
		WriteTimeout: s.cfg.HTTPServer.Timeout,
		IdleTimeout:  s.cfg.HTTPServer.IdleTimeout,
	}
	return s.server.ListenAndServe()
}

// Stop останавливает сервер с graceful shutdown
func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
