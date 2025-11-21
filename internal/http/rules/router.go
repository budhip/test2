package rules

import (
	"bitbucket.org/Amartha/go-megatron/internal/repositories"

	"github.com/labstack/echo/v4"
)

// RegisterRoutes registers all rule management routes
func RegisterRoutes(e *echo.Group, repo repositories.RuleRepository) {
	handler := NewHandler(repo)

	rules := e.Group("/rules")

	// POST /api/v1/rules - Create a new rule
	rules.POST("", handler.CreateRule)

	// GET /api/v1/rules?env=dev - List all rules for an environment
	rules.GET("", handler.ListRules)

	// GET /api/v1/rules/:name?env=dev&version=0.0.1 - Get a specific rule
	rules.GET("/:name", handler.GetRule)

	// PUT /api/v1/rules/:id - Update a rule
	rules.PUT("/:id", handler.UpdateRule)

	// PATCH /api/v1/rules/:id/append - Append new rule to existing content
	rules.PATCH("/:id/append", handler.AppendRule)

	// DELETE /api/v1/rules/:id - Delete a rule
	rules.DELETE("/:id", handler.DeleteRule)
}
