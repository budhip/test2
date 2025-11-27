package rule

import (
	"net/http"

	ruleCmd "bitbucket.org/Amartha/go-megatron/internal/application/command/rule"
	"bitbucket.org/Amartha/go-megatron/internal/interfaces/http/dto/request"
	"bitbucket.org/Amartha/go-megatron/internal/interfaces/http/dto/response"

	"github.com/labstack/echo/v4"
)

type UpdateRuleHTTPHandler struct {
	updateRuleHandler *ruleCmd.UpdateRuleHandler
}

func NewUpdateRuleHTTPHandler(handler *ruleCmd.UpdateRuleHandler) *UpdateRuleHTTPHandler {
	return &UpdateRuleHTTPHandler{
		updateRuleHandler: handler,
	}
}

func (h *UpdateRuleHTTPHandler) Handle(c echo.Context) error {
	ctx := c.Request().Context()
	ruleID := c.Param("id")

	var req request.UpdateRuleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
	}

	if err := req.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
	}

	cmd := ruleCmd.UpdateRuleCommand{
		RuleID:  ruleID,
		Content: req.Content,
	}

	if err := h.updateRuleHandler.Handle(ctx, cmd); err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "Failed to update rule",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response.MessageResponse{
		Message: "Rule updated successfully",
	})
}
