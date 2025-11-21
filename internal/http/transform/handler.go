package transform

import (
	"net/http"

	xlog "bitbucket.org/Amartha/go-x/log"

	"bitbucket.org/Amartha/go-megatron/internal/models"
	"bitbucket.org/Amartha/go-megatron/internal/services"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	transformService services.TransformService
}

func NewHandler(transformService services.TransformService) *Handler {
	return &Handler{
		transformService: transformService,
	}
}

// TransformWalletTransaction handles wallet transaction transformation using Grule
// POST /api/v1/transform/wallet
func (h *Handler) TransformWalletTransaction(c echo.Context) error {
	ctx := c.Request().Context()
	var req models.WalletTransactionRequest

	if err := c.Bind(&req); err != nil {
		xlog.Warn(ctx, "[TRANSFORM_HANDLER] Invalid request body", xlog.Err(err))
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
	}

	// Validate request
	if req.WalletTransaction.ID == "" {
		xlog.Warn(ctx, "[TRANSFORM_HANDLER] Missing wallet transaction ID")
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Validation failed",
			Message: "wallet transaction ID is required",
		})
	}

	if req.WalletTransaction.TransactionType == "" {
		xlog.Warn(ctx, "[TRANSFORM_HANDLER] Missing transaction type")
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Validation failed",
			Message: "transaction type is required",
		})
	}

	// Log request
	xlog.Info(ctx, "[TRANSFORM_HANDLER] Wallet transaction transform request received",
		xlog.String("wallet_transaction_id", req.WalletTransaction.ID),
		xlog.String("transaction_type", req.WalletTransaction.TransactionType),
		xlog.String("account_number", req.WalletTransaction.AccountNumber),
		xlog.String("net_amount", req.WalletTransaction.NetAmount.Value.String()),
		xlog.Int("amounts_count", len(req.WalletTransaction.Amounts)))

	// Execute transformation
	resp, err := h.transformService.TransformWalletTransaction(ctx, req)
	if err != nil {
		xlog.Error(ctx, "[TRANSFORM_HANDLER] Wallet transaction transform failed",
			xlog.Err(err),
			xlog.String("wallet_transaction_id", req.WalletTransaction.ID))
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Wallet transaction transform failed",
			Message: err.Error(),
		})
	}

	// Log success
	xlog.Info(ctx, "[TRANSFORM_HANDLER] Wallet transaction transform success",
		xlog.String("wallet_transaction_id", req.WalletTransaction.ID),
		xlog.Int("transaction_count", len(resp.Transactions)),
		xlog.Int("execution_time_ms", resp.Metadata.ExecutionTimeMs))

	return c.JSON(http.StatusOK, resp)
}
