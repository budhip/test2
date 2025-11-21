package transform

import (
	"bitbucket.org/Amartha/go-megatron/internal/services"

	"github.com/labstack/echo/v4"
)

// RegisterRoutes registers all transformation routes
func RegisterRoutes(e *echo.Group, transformService services.TransformService) {
	handler := NewHandler(transformService)

	// POST /api/v1/transform/wallet - Transform wallet transaction
	e.POST("/transform/wallet", handler.TransformWalletTransaction)
}
