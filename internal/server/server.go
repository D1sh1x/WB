package server

import (
	"context"
	"log/slog"
	"net/http"

	"WB2/internal/config"
	"WB2/internal/kafka"
	storage "WB2/internal/storage/postgres"

	"github.com/labstack/echo/v4"
)

type Server struct {
	cfg    *config.Config
	router *echo.Echo
	server *http.Server
	kafka  *kafka.Service
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		cfg:    cfg,
		router: echo.New(),
	}
}

func (s *Server) Start(log *slog.Logger, storage *storage.Storage, kafka *kafka.Service) error {
	s.kafka = kafka
	InitRoutes(s.router, log, storage, s.cfg, kafka)
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
