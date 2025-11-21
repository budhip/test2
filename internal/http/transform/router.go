package transform

import (
	"bitbucket.org/Amartha/go-megatron/internal/service"
	"github.com/labstack/echo/v4"
)

// RegisterRoutes registers all transformation routes
func RegisterRoutes(e *echo.Group, transformService service.TransformService) {
	handler := NewHandler(transformService)

	// POST /api/v1/transform - Transform single transaction
	e.POST("/transform", handler.Transform)

	// POST /api/v1/transform/batch - Transform multiple transactions
	e.POST("/transform/batch", handler.BatchTransform)
}
