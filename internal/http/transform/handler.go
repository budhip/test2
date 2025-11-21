package transform

import (
	"net/http"

	"bitbucket.org/Amartha/go-megatron/internal/megatron"
	"bitbucket.org/Amartha/go-megatron/internal/service"
	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	transformService service.TransformService
}

func NewHandler(transformService service.TransformService) *Handler {
	return &Handler{
		transformService: transformService,
	}
}

// Transform handles single transformation request
func (h *Handler) Transform(c echo.Context) error {
	ctx := c.Request().Context()
	var req megatron.TransformRequest

	if err := c.Bind(&req); err != nil {
		xlog.Warn(ctx, "[TRANSFORM_HANDLER] Invalid request body", xlog.Err(err))
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
	}

	xlog.Info(ctx, "[TRANSFORM_HANDLER] Transform request received",
		xlog.String("transaction_type", req.TransactionType),
		xlog.String("wallet_transaction_id", req.ParentTransaction.ID))

	resp, err := h.transformService.Transform(ctx, req)
	if err != nil {
		xlog.Error(ctx, "[TRANSFORM_HANDLER] Transform failed", xlog.Err(err))
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Transform failed",
			Message: err.Error(),
		})
	}

	xlog.Info(ctx, "[TRANSFORM_HANDLER] Transform success",
		xlog.String("transaction_type", req.TransactionType),
		xlog.Int("transaction_count", len(resp.Transactions)),
		xlog.Int("execution_time_ms", resp.Metadata.ExecutionTimeMs))

	return c.JSON(http.StatusOK, resp)
}

// BatchTransform handles batch transformation request
func (h *Handler) BatchTransform(c echo.Context) error {
	ctx := c.Request().Context()
	var req megatron.BatchTransformRequest

	if err := c.Bind(&req); err != nil {
		xlog.Warn(ctx, "[TRANSFORM_HANDLER] Invalid request body", xlog.Err(err))
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
	}

	xlog.Info(ctx, "[TRANSFORM_HANDLER] Batch transform request received",
		xlog.String("wallet_transaction_id", req.ParentTransaction.ID),
		xlog.Int("transform_count", len(req.Transforms)))

	resp, err := h.transformService.BatchTransform(ctx, req)
	if err != nil {
		xlog.Error(ctx, "[TRANSFORM_HANDLER] Batch transform failed", xlog.Err(err))
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Batch transform failed",
			Message: err.Error(),
		})
	}

	xlog.Info(ctx, "[TRANSFORM_HANDLER] Batch transform success",
		xlog.Int("transaction_count", len(resp.Transactions)),
		xlog.Int("error_count", len(resp.Errors)),
		xlog.Int("execution_time_ms", resp.Metadata.ExecutionTimeMs))

	return c.JSON(http.StatusOK, resp)
}
