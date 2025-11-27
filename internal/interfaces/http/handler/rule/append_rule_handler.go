package rule

import (
	"net/http"

	ruleCmd "bitbucket.org/Amartha/go-megatron/internal/application/command/rule"
	"bitbucket.org/Amartha/go-megatron/internal/interfaces/http/dto/request"
	"bitbucket.org/Amartha/go-megatron/internal/interfaces/http/dto/response"

	"github.com/labstack/echo/v4"
)

type AppendRuleHTTPHandler struct {
	appendRuleHandler *ruleCmd.AppendRuleHandler
}

func NewAppendRuleHTTPHandler(handler *ruleCmd.AppendRuleHandler) *AppendRuleHTTPHandler {
	return &AppendRuleHTTPHandler{
		appendRuleHandler: handler,
	}
}

func (h *AppendRuleHTTPHandler) Handle(c echo.Context) error {
	ctx := c.Request().Context()
	ruleID := c.Param("id")

	var req request.AppendRuleRequest
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

	if req.InsertMode == "" {
		req.InsertMode = "end"
	}
	if req.VersionBump == "" && req.AutoVersion {
		req.VersionBump = "patch"
	}

	cmd := ruleCmd.AppendRuleCommand{
		RuleID:      ruleID,
		Content:     req.Content,
		InsertMode:  req.InsertMode,
		AutoVersion: req.AutoVersion,
		VersionBump: req.VersionBump,
	}

	if err := h.appendRuleHandler.Handle(ctx, cmd); err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "Failed to append rule",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response.MessageResponse{
		Message: "Rule appended successfully",
	})
}
