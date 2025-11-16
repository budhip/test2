package http

import (
	"context"
	"fmt"

	"bitbucket.org/Amartha/go-megatron/internal/config"
	rulesHttp "bitbucket.org/Amartha/go-megatron/internal/http/rules"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/graceful"
	"bitbucket.org/Amartha/go-megatron/internal/repository"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type httpAPI struct {
	e        *echo.Echo
	cfg      *config.Configuration
	ruleRepo repository.RuleRepository
}

// NewAPI creates a new HTTP API server
func NewAPI(cfg *config.Configuration, ruleRepo repository.RuleRepository) Http {
	return &httpAPI{
		e:        echo.New(),
		cfg:      cfg,
		ruleRepo: ruleRepo,
	}
}

func (h *httpAPI) Start() (graceful.ProcessStarter, graceful.ProcessStopper) {
	// Middleware
	h.e.Use(middleware.Logger())
	h.e.Use(middleware.Recover())
	h.e.Use(middleware.CORS())
	h.e.Use(middleware.RequestID())

	// Custom middleware to add request ID to response header
	h.e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqID := c.Response().Header().Get(echo.HeaderXRequestID)
			c.Response().Header().Set("X-Request-ID", reqID)
			return next(c)
		}
	})

	// Root endpoint
	h.e.GET("/", func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"service": "Go Megatron API Server",
			"version": "1.0.0",
			"env":     h.cfg.App.Env,
		})
	})

	// Health check endpoint
	h.e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "healthy",
		})
	})

	// API v1 routes
	v1 := h.e.Group("/api/v1")

	// Register rules management routes
	rulesHttp.RegisterRoutes(v1, h.ruleRepo)

	// Determine port based on environment
	port := ":8080"
	if h.cfg.App.Env == "prod" {
		port = ":80"
	}

	return func() error {
			fmt.Printf("🚀 API Server starting on %s (env: %s)\n", port, h.cfg.App.Env)
			return h.e.Start(port)
		},
		func(ctx context.Context) error {
			fmt.Println("⏳ Shutting down API server...")
			return h.e.Shutdown(ctx)
		}
}
