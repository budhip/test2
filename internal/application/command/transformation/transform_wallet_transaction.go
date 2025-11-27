package transformation

import (
	"context"
	"fmt"
	"time"

	"bitbucket.org/Amartha/go-megatron/internal/application/dto"
	"bitbucket.org/Amartha/go-megatron/internal/domain/transformation"

	"github.com/shopspring/decimal"
)

type TransformWalletTransactionCommand struct {
	WalletTransactionID      string
	Status                   string
	AccountNumber            string
	DestinationAccountNumber string
	RefNumber                string
	TransactionType          string
	TransactionTime          time.Time
	TransactionFlow          string

	NetAmount struct {
		Value    decimal.Decimal
		Currency string
	}

	Amounts []struct {
		Type   string
		Amount struct {
			Value    decimal.Decimal
			Currency string
		}
	}

	Description string
	Metadata    map[string]interface{}
	CreatedAt   time.Time
}

type TransformWalletTransactionHandler struct {
	transformationService *transformation.TransformationService
	transformationRepo    transformation.TransformationRepository
}

func NewTransformWalletTransactionHandler(
	service *transformation.TransformationService,
	repo transformation.TransformationRepository,
) *TransformWalletTransactionHandler {
	return &TransformWalletTransactionHandler{
		transformationService: service,
		transformationRepo:    repo,
	}
}

func (h *TransformWalletTransactionHandler) Handle(
	ctx context.Context,
	cmd TransformWalletTransactionCommand,
) (*dto.TransformationResultDTO, error) {

	startTime := time.Now()

	walletTx, err := transformation.NewWalletTransaction(
		cmd.WalletTransactionID,
		cmd.Status,
		cmd.AccountNumber,
		cmd.DestinationAccountNumber,
		cmd.RefNumber,
		cmd.TransactionType,
		cmd.TransactionTime,
		cmd.TransactionFlow,
		cmd.NetAmount.Value,
		cmd.NetAmount.Currency,
		cmd.Description,
		cmd.Metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet transaction: %w", err)
	}

	for _, amountBreakdown := range cmd.Amounts {
		if err := walletTx.AddAmountBreakdown(
			amountBreakdown.Type,
			amountBreakdown.Amount.Value,
			amountBreakdown.Amount.Currency,
		); err != nil {
			return nil, fmt.Errorf("failed to add amount breakdown: %w", err)
		}
	}

	transactions, err := h.transformationService.TransformWalletTransaction(ctx, walletTx)
	if err != nil {
		return nil, fmt.Errorf("transformation failed: %w", err)
	}

	if err := h.transformationRepo.SaveWalletTransaction(ctx, walletTx); err != nil {
		return nil, fmt.Errorf("failed to save wallet transaction: %w", err)
	}

	if err := h.transformationRepo.SaveTransactions(ctx, transactions); err != nil {
		return nil, fmt.Errorf("failed to save transactions: %w", err)
	}

	executionTime := time.Since(startTime)

	return dto.NewTransformationResultDTO(
		walletTx,
		transactions,
		executionTime,
	), nil
}
