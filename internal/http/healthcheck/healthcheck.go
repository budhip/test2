package healthcheck

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type HealthCheck struct{}

func NewHealthCheck() *HealthCheck {
	return &HealthCheck{}
}

func (d *HealthCheck) Route(g *echo.Group) {
	g.GET("/consumer", d.healthCheckConsumerHandler)
}

// @Summary Health Check Consumer
// @Description Checking consumer service health
// @Tags Health
// @Produce json
// @Success 200 {object} response.HealthCheckSuccessModel "Success"
// @Router /health/consumer [get]
func (d *HealthCheck) healthCheckConsumerHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, nil)
}
