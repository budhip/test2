package transformation

import (
	"net/http"

	transformCmd "bitbucket.org/Amartha/go-megatron/internal/application/command/transformation"
	"bitbucket.org/Amartha/go-megatron/internal/interfaces/http/dto/request"
	"bitbucket.org/Amartha/go-megatron/internal/interfaces/http/dto/response"

	"github.com/labstack/echo/v4"

	"github.com/shopspring/decimal"
)

type TransformWalletHTTPHandler struct {
	transformHandler *transformCmd.TransformWalletTransactionHandler
}

func NewTransformWalletHTTPHandler(
	handler *transformCmd.TransformWalletTransactionHandler,
) *TransformWalletHTTPHandler {
	return &TransformWalletHTTPHandler{
		transformHandler: handler,
	}
}

func (h *TransformWalletHTTPHandler) Handle(c echo.Context) error {
	ctx := c.Request().Context()

	var req request.TransformWalletTransactionRequest
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

	cmd := transformCmd.TransformWalletTransactionCommand{
		WalletTransactionID:      req.WalletTransaction.ID,
		Status:                   req.WalletTransaction.Status,
		AccountNumber:            req.WalletTransaction.AccountNumber,
		DestinationAccountNumber: req.WalletTransaction.DestinationAccountNumber,
		RefNumber:                req.WalletTransaction.RefNumber,
		TransactionType:          req.WalletTransaction.TransactionType,
		TransactionTime:          req.WalletTransaction.TransactionTime,
		TransactionFlow:          req.WalletTransaction.TransactionFlow,
		Description:              req.WalletTransaction.Description,
		Metadata:                 req.WalletTransaction.Metadata,
		CreatedAt:                req.WalletTransaction.CreatedAt,
	}

	cmd.NetAmount.Value = req.WalletTransaction.NetAmount.Value
	cmd.NetAmount.Currency = req.WalletTransaction.NetAmount.Currency

	for _, amt := range req.WalletTransaction.Amounts {
		cmd.Amounts = append(cmd.Amounts, struct {
			Type   string
			Amount struct {
				Value    decimal.Decimal
				Currency string
			}
		}{
			Type: amt.Type,
			Amount: struct {
				Value    decimal.Decimal
				Currency string
			}{
				Value:    amt.Amount.Value,
				Currency: amt.Amount.Currency,
			},
		})
	}

	result, err := h.transformHandler.Handle(ctx, cmd)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error:   "Transformation failed",
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response.FromTransformationResultDTO(result))
}
