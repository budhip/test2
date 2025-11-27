package rule

import (
	"net/http"

	ruleCmd "bitbucket.org/Amartha/go-megatron/internal/application/command/rule"
	"bitbucket.org/Amartha/go-megatron/internal/interfaces/http/dto/request"
	"bitbucket.org/Amartha/go-megatron/internal/interfaces/http/dto/response"

	"github.com/labstack/echo/v4"
)

type CreateRuleHTTPHandler struct {
	createRuleHandler *ruleCmd.CreateRuleHandler
}

func NewCreateRuleHTTPHandler(handler *ruleCmd.CreateRuleHandler) *CreateRuleHTTPHandler {
	return &CreateRuleHTTPHandler{
		createRuleHandler: handler,
	}
}

func (h *CreateRuleHTTPHandler) Handle(c echo.Context) error {
	ctx := c.Request().Context()

	var req request.CreateRuleRequest
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

	cmd := ruleCmd.CreateRuleCommand{
		Name:    req.Name,
		Env:     req.Env,
		Version: req.Version,
		Content: req.Content,
	}

	result, err := h.createRuleHandler.Handle(ctx, cmd)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "Failed to create rule",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, response.FromRuleDTO(result))
}
