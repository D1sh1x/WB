package server

import (
	"WB2/internal/config"
	storage "WB2/internal/storage/postgres"
	"context"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Server struct {
	cfg    *config.Config
	router *echo.Echo
	server *http.Server
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		cfg:    cfg,
		router: echo.New(),
	}
}

func (s *Server) Start(log *slog.Logger, storage *storage.Storage) error {
	InitRoutes(s.router, log, storage, s.cfg)
	s.server = &http.Server{
		Addr:         ":" + s.cfg.HTTPServer.Port,
		Handler:      s.router,
		ReadTimeout:  s.cfg.HTTPServer.Timeout,
		WriteTimeout: s.cfg.HTTPServer.Timeout,
		IdleTimeout:  s.cfg.HTTPServer.IdleTimeout,
	}
	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
