package rule

import (
	"net/http"

	ruleQuery "bitbucket.org/Amartha/go-megatron/internal/application/query/rule"
	"bitbucket.org/Amartha/go-megatron/internal/interfaces/http/dto/response"

	"github.com/labstack/echo/v4"
)

type GetRuleHTTPHandler struct {
	getRuleHandler *ruleQuery.GetRuleHandler
}

func NewGetRuleHTTPHandler(handler *ruleQuery.GetRuleHandler) *GetRuleHTTPHandler {
	return &GetRuleHTTPHandler{
		getRuleHandler: handler,
	}
}

func (h *GetRuleHTTPHandler) Handle(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.Param("name")
	env := c.QueryParam("env")
	version := c.QueryParam("version")

	if name == "" || env == "" || version == "" {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "Missing parameters",
			Message: "name, env, and version are required",
		})
	}

	query := ruleQuery.GetRuleQuery{
		Name:    name,
		Env:     env,
		Version: version,
	}

	result, err := h.getRuleHandler.Handle(ctx, query)
	if err != nil {
		return c.JSON(http.StatusNotFound, response.ErrorResponse{
			Error:   "Rule not found",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response.FromRuleDTO(result))
}
