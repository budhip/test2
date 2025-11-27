package rule

import (
	"net/http"

	ruleQuery "bitbucket.org/Amartha/go-megatron/internal/application/query/rule"
	"bitbucket.org/Amartha/go-megatron/internal/interfaces/http/dto/response"

	"github.com/labstack/echo/v4"
)

type ListRulesHTTPHandler struct {
	listRulesHandler *ruleQuery.ListRulesHandler
}

func NewListRulesHTTPHandler(handler *ruleQuery.ListRulesHandler) *ListRulesHTTPHandler {
	return &ListRulesHTTPHandler{
		listRulesHandler: handler,
	}
}

func (h *ListRulesHTTPHandler) Handle(c echo.Context) error {
	ctx := c.Request().Context()

	env := c.QueryParam("env")
	if env == "" {
		return c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error:   "Missing parameter",
			Message: "env is required",
		})
	}

	query := ruleQuery.ListRulesQuery{
		Env: env,
	}

	results, err := h.listRulesHandler.Handle(ctx, query)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "Failed to list rules",
			Message: err.Error(),
		})
	}

	responseList := make([]response.RuleResponse, len(results))
	for i, r := range results {
		responseList[i] = response.FromRuleDTO(r)
	}

	return c.JSON(http.StatusOK, responseList)
}
