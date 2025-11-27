package dto

import (
	"time"

	"bitbucket.org/Amartha/go-megatron/internal/domain/transformation"

	"github.com/shopspring/decimal"
)

type TransformationResultDTO struct {
	Transactions []TransactionDTO      `json:"transactions"`
	Metadata     TransformationMetaDTO `json:"metadata"`
	Warnings     []string              `json:"warnings,omitempty"`
}

type TransactionDTO struct {
	TransactionID   string                 `json:"transactionId"`
	FromAccount     string                 `json:"fromAccount"`
	ToAccount       string                 `json:"toAccount"`
	FromNarrative   string                 `json:"fromNarrative,omitempty"`
	ToNarrative     string                 `json:"toNarrative,omitempty"`
	TransactionDate string                 `json:"transactionDate"`
	Amount          decimal.Decimal        `json:"amount"`
	Currency        string                 `json:"currency"`
	Status          string                 `json:"status"`
	TypeTransaction string                 `json:"typeTransaction"`
	Description     string                 `json:"description"`
	RefNumber       string                 `json:"refNumber"`
	OrderType       string                 `json:"orderType"`
	OrderTime       time.Time              `json:"orderTime"`
	TransactionTime time.Time              `json:"transactionTime"`
	Metadata        map[string]interface{} `json:"metadata"`
}

type TransformationMetaDTO struct {
	ExecutionTimeMs  int    `json:"executionTimeMs"`
	TransactionCount int    `json:"transactionCount"`
	ProcessedAt      string `json:"processedAt"`
}

func NewTransformationResultDTO(
	wt *transformation.WalletTransaction,
	transactions []*transformation.Transaction,
	executionTime time.Duration,
) *TransformationResultDTO {

	txDTOs := make([]TransactionDTO, len(transactions))
	for i, tx := range transactions {
		txDTOs[i] = TransactionDTO{
			TransactionID:   tx.ID().String(),
			FromAccount:     tx.FromAccount().String(),
			ToAccount:       tx.ToAccount().String(),
			FromNarrative:   tx.FromNarrative(),
			ToNarrative:     tx.ToNarrative(),
			TransactionDate: tx.TransactionDate().Format(),
			Amount:          tx.Amount().Value(),
			Currency:        tx.Amount().Currency().Code(),
			Status:          tx.Status().String(),
			TypeTransaction: tx.TransactionType().String(),
			Description:     tx.Description(),
			RefNumber:       tx.RefNumber().String(),
			OrderType:       tx.OrderType().String(),
			OrderTime:       tx.OrderTime(),
			TransactionTime: tx.TransactionTime(),
			Metadata:        tx.Metadata().ToMap(),
		}
	}

	return &TransformationResultDTO{
		Transactions: txDTOs,
		Metadata: TransformationMetaDTO{
			ExecutionTimeMs:  int(executionTime.Milliseconds()),
			TransactionCount: len(transactions),
			ProcessedAt:      time.Now().Format(time.RFC3339),
		},
	}
}
