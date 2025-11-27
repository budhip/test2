package response

import (
	"time"

	"bitbucket.org/Amartha/go-megatron/internal/application/dto"

	"github.com/shopspring/decimal"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type RuleResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Env       string    `json:"env"`
	Version   string    `json:"version"`
	Content   string    `json:"content"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func FromRuleDTO(dto *dto.RuleDTO) RuleResponse {
	return RuleResponse{
		ID:        dto.ID,
		Name:      dto.Name,
		Env:       dto.Env,
		Version:   dto.Version,
		Content:   dto.Content,
		IsActive:  dto.IsActive,
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
	}
}

type TransformationResultResponse struct {
	Transactions []TransactionResponse      `json:"transactions"`
	Metadata     TransformationMetaResponse `json:"metadata"`
	Warnings     []string                   `json:"warnings,omitempty"`
}

type TransactionResponse struct {
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

type TransformationMetaResponse struct {
	ExecutionTimeMs  int    `json:"executionTimeMs"`
	TransactionCount int    `json:"transactionCount"`
	ProcessedAt      string `json:"processedAt"`
}

func FromTransformationResultDTO(dto *dto.TransformationResultDTO) TransformationResultResponse {
	transactions := make([]TransactionResponse, len(dto.Transactions))
	for i, tx := range dto.Transactions {
		transactions[i] = TransactionResponse{
			TransactionID:   tx.TransactionID,
			FromAccount:     tx.FromAccount,
			ToAccount:       tx.ToAccount,
			FromNarrative:   tx.FromNarrative,
			ToNarrative:     tx.ToNarrative,
			TransactionDate: tx.TransactionDate,
			Amount:          tx.Amount,
			Currency:        tx.Currency,
			Status:          tx.Status,
			TypeTransaction: tx.TypeTransaction,
			Description:     tx.Description,
			RefNumber:       tx.RefNumber,
			OrderType:       tx.OrderType,
			OrderTime:       tx.OrderTime,
			TransactionTime: tx.TransactionTime,
			Metadata:        tx.Metadata,
		}
	}

	return TransformationResultResponse{
		Transactions: transactions,
		Metadata: TransformationMetaResponse{
			ExecutionTimeMs:  dto.Metadata.ExecutionTimeMs,
			TransactionCount: dto.Metadata.TransactionCount,
			ProcessedAt:      dto.Metadata.ProcessedAt,
		},
		Warnings: dto.Warnings,
	}
}
