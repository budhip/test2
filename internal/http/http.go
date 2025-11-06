package http

import (
	"bitbucket.org/Amartha/go-megatron/internal/pkg/graceful"
	"github.com/labstack/echo/v4"
)

type (
	Http interface {
		Start() (graceful.ProcessStarter, graceful.ProcessStopper)
	}
	httpApi struct {
		e *echo.Echo
	}
)
