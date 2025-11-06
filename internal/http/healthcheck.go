package http

import (
	"context"

	"bitbucket.org/Amartha/go-megatron/internal/config"
	"bitbucket.org/Amartha/go-megatron/internal/http/healthcheck"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/graceful"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
)

type httpHealthCheck struct {
	e   *echo.Echo
	cfg *config.Configuration
}

func NewHealthCheck(cfg *config.Configuration) Http {
	return &httpHealthCheck{
		e:   echo.New(),
		cfg: cfg,
	}
}

func (h *httpHealthCheck) Start() (graceful.ProcessStarter, graceful.ProcessStopper) {
	// metrics
	h.e.GET("/metrics", echoprometheus.NewHandler())

	healthcheck.Route(h.e.Group("/health"))
	return func() error {
			return h.e.Start(":" + h.cfg.Kafka.Consumers.HealthCheckPort)
		},
		func(ctx context.Context) error {
			return h.e.Shutdown(ctx)
		}
}
