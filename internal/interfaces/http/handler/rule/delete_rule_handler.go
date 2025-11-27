package rule

import (
	"net/http"

	ruleCmd "bitbucket.org/Amartha/go-megatron/internal/application/command/rule"
	"bitbucket.org/Amartha/go-megatron/internal/interfaces/http/dto/response"

	"github.com/labstack/echo/v4"
)

type DeleteRuleHTTPHandler struct {
	deleteRuleHandler *ruleCmd.DeleteRuleHandler
}

func NewDeleteRuleHTTPHandler(handler *ruleCmd.DeleteRuleHandler) *DeleteRuleHTTPHandler {
	return &DeleteRuleHTTPHandler{
		deleteRuleHandler: handler,
	}
}

func (h *DeleteRuleHTTPHandler) Handle(c echo.Context) error {
	ctx := c.Request().Context()
	ruleID := c.Param("id")

	cmd := ruleCmd.DeleteRuleCommand{
		RuleID: ruleID,
	}

	if err := h.deleteRuleHandler.Handle(ctx, cmd); err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "Failed to delete rule",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response.MessageResponse{
		Message: "Rule deleted successfully",
	})
}
