package http

import (
	"bitbucket.org/Amartha/go-megatron/internal/acuanrepository"
	"context"
	"fmt"

	"bitbucket.org/Amartha/go-megatron/internal/config"
	rulesHttp "bitbucket.org/Amartha/go-megatron/internal/http/rules"
	transformHttp "bitbucket.org/Amartha/go-megatron/internal/http/transform"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/graceful"
	"bitbucket.org/Amartha/go-megatron/internal/repository"
	"bitbucket.org/Amartha/go-megatron/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type httpAPI struct {
	e                *echo.Echo
	cfg              *config.Configuration
	ruleRepo         repository.RuleRepository
	acuanRuleRepo    acuanrepository.RuleRepository
	transformService service.TransformService
}

// NewAPI creates a new HTTP API server
func NewAPI(cfg *config.Configuration, acuanRuleRepo acuanrepository.RuleRepository, ruleRepo repository.RuleRepository) Http {
	// Create transform service
	transformService := service.NewTransformService(acuanRuleRepo)

	return &httpAPI{
		e:                echo.New(),
		cfg:              cfg,
		ruleRepo:         ruleRepo,
		acuanRuleRepo:    acuanRuleRepo,
		transformService: transformService,
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
			"endpoints": map[string][]string{
				"transform": {
					"POST /api/v1/transform",
					"POST /api/v1/transform/batch",
				},
				"rules": {
					"POST /api/v1/rules",
					"GET /api/v1/rules",
					"GET /api/v1/rules/:name",
					"PUT /api/v1/rules/:id",
					"PATCH /api/v1/rules/:id/append",
					"DELETE /api/v1/rules/:id",
				},
			},
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

	// Register transformation routes
	transformHttp.RegisterRoutes(v1, h.transformService)

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
