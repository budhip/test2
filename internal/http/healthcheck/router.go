package healthcheck

import (
	"github.com/labstack/echo/v4"
)

func Route(g *echo.Group) {
	NewHealthCheck().Route(g)
}
