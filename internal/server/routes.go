package server

import (
	"log/slog"

	"WB2/internal/cache"
	"WB2/internal/config"
	"WB2/internal/handler"
	storage "WB2/internal/storage/postgres"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func InitRoutes(router *echo.Echo, log *slog.Logger, storage *storage.Storage, cfg *config.Config, c *cache.OrderCache) {

	router.Use(middleware.Logger())
	router.Use(middleware.Recover())
	router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowMethods: []string{"GET", "HEAD", "PUT", "PATCH", "POST", "DELETE"},
	}))

	h := handler.NewHandler(log, storage, c)

	router.GET("/order", h.GetAllOrdres)
	router.GET("/order/:id", h.GetOrderByID)
	router.POST("/order", h.CreateOrder)
	router.PUT("/order", h.UpdateOrder)
	router.DELETE("/order", h.DeleteOrder)

	// healthcheck
	router.GET("/health", func(c echo.Context) error { return c.JSON(200, map[string]string{"status": "ok"}) })
}
