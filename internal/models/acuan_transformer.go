package models

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

// WalletTransactionRequest adalah request untuk transform wallet transaction ke acuan
type WalletTransactionRequest struct {
	WalletTransaction WalletTransaction `json:"walletTransaction"`
	Context           TransformContext  `json:"context"`
}

type WalletTransaction struct {
	ID                       string                 `json:"id"`
	Status                   string                 `json:"status"`
	AccountNumber            string                 `json:"accountNumber"`
	DestinationAccountNumber string                 `json:"destinationAccountNumber"`
	RefNumber                string                 `json:"refNumber"`
	TransactionType          string                 `json:"transactionType"`
	TransactionTime          time.Time              `json:"transactionTime"`
	TransactionFlow          string                 `json:"transactionFlow"`
	NetAmount                Amount                 `json:"netAmount"`
	Amounts                  []AmountWithType       `json:"amounts"`
	Description              string                 `json:"description"`
	Metadata                 map[string]interface{} `json:"metadata"`
	CreatedAt                time.Time              `json:"createdAt"`
}

type Amount struct {
	Value    decimal.Decimal `json:"value"`
	Currency string          `json:"currency"`
}

type AmountWithType struct {
	Type   string `json:"type"`
	Amount Amount `json:"amount"`
}

type WalletTransactionResponse struct {
	Transactions []TransactionReq  `json:"transactions"`
	Metadata     TransformMetadata `json:"metadata"`
	Warnings     []string          `json:"warnings,omitempty"`
}

type TransactionReq struct {
	TransactionID   string                 `json:"transactionId"`
	FromAccount     string                 `json:"fromAccount"`
	ToAccount       string                 `json:"toAccount"`
	FromNarrative   string                 `json:"fromNarrative"`
	ToNarrative     string                 `json:"toNarrative"`
	TransactionDate string                 `json:"transactionDate"`
	Amount          decimal.Decimal        `json:"amount"`
	Status          string                 `json:"status"`
	Method          string                 `json:"method"`
	TypeTransaction string                 `json:"typeTransaction"`
	Description     string                 `json:"description"`
	RefNumber       string                 `json:"refNumber"`
	Metadata        map[string]interface{} `json:"metadata"`
	OrderTime       time.Time              `json:"orderTime"`
	OrderType       string                 `json:"orderType"`
	TransactionTime time.Time              `json:"transactionTime"`
	Currency        string                 `json:"currency"`
}

// TransformRequest adalah request untuk transform wallet transaction
type TransformRequest struct {
	// Parent wallet transaction info
	ParentTransaction WalletTransactionInput `json:"parentTransaction"`

	// Amount to transform (bisa dari NetAmount atau dari Amounts)
	Amount AmountInput `json:"amount"`

	// Transaction type yang akan di-transform
	TransactionType string `json:"transactionType"`

	// Context information
	Context TransformContext `json:"context"`
}

type WalletTransactionInput struct {
	ID                       string                 `json:"id"`
	AccountNumber            string                 `json:"accountNumber"`
	DestinationAccountNumber string                 `json:"destinationAccountNumber"`
	RefNumber                string                 `json:"refNumber"`
	TransactionType          string                 `json:"transactionType"`
	TransactionFlow          string                 `json:"transactionFlow"`
	TransactionTime          time.Time              `json:"transactionTime"`
	Description              string                 `json:"description"`
	Metadata                 map[string]interface{} `json:"metadata"`
	Status                   string                 `json:"status"`

	// Account info (akan di-populate dari database)
	Account AccountInfo `json:"account,omitempty"`
}

type AccountInfo struct {
	AccountNumber   string `json:"accountNumber"`
	Name            string `json:"name"`
	Entity          string `json:"entity"`
	CategoryCode    string `json:"categoryCode"`
	SubCategoryCode string `json:"subCategoryCode"`
}

type AmountInput struct {
	Value    decimal.Decimal `json:"value"`
	Currency string          `json:"currency"`
}

type TransformContext struct {
	// Config yang dibutuhkan (akan di-inject dari go-fp-transaction)
	SystemAccountNumber                               string            `json:"systemAccountNumber"`
	AccountNumberInsurancePremiumDisbursementByEntity map[string]string `json:"accountNumberInsurancePremiumDisbursementByEntity,omitempty"`
	MapAccountEntity                                  map[string]string `json:"mapAccountEntity,omitempty"`

	// Additional context
	ClientID      string `json:"clientId,omitempty"`
	CorrelationID string `json:"correlationId,omitempty"`
	RequestID     string `json:"requestId,omitempty"`
}

// TransformResponse adalah response dari transformation
type TransformResponse struct {
	// Hasil transformasi
	Transactions []TransactionOutput `json:"transactions"`

	// Metadata
	Metadata TransformMetadata `json:"metadata"`

	// Validation warnings (jika ada)
	Warnings []string `json:"warnings,omitempty"`
}

type TransactionOutput struct {
	TransactionID   string                 `json:"transactionId"`
	FromAccount     string                 `json:"fromAccount"`
	ToAccount       string                 `json:"toAccount"`
	FromNarrative   string                 `json:"fromNarrative"`
	ToNarrative     string                 `json:"toNarrative"`
	TransactionDate string                 `json:"transactionDate"`
	Amount          decimal.Decimal        `json:"amount"`
	Status          string                 `json:"status"`
	Method          string                 `json:"method"`
	TypeTransaction string                 `json:"typeTransaction"`
	Description     string                 `json:"description"`
	RefNumber       string                 `json:"refNumber"`
	Metadata        map[string]interface{} `json:"metadata"`
	OrderTime       time.Time              `json:"orderTime"`
	OrderType       string                 `json:"orderType"`
	TransactionTime time.Time              `json:"transactionTime"`
	Currency        string                 `json:"currency"`
}

type TransformMetadata struct {
	RuleID          string `json:"ruleId"`
	RuleName        string `json:"ruleName"`
	RuleVersion     int    `json:"ruleVersion"`
	ExecutionTimeMs int    `json:"executionTimeMs"`
}

// BatchTransformRequest untuk transform multiple amounts sekaligus
type BatchTransformRequest struct {
	ParentTransaction WalletTransactionInput `json:"parentTransaction"`
	Transforms        []TransformItem        `json:"transforms"`
	Context           TransformContext       `json:"context"`
}

type TransformItem struct {
	Amount          AmountInput `json:"amount"`
	TransactionType string      `json:"transactionType"`
}

type BatchTransformResponse struct {
	Transactions []TransactionOutput `json:"transactions"`
	Errors       []TransformError    `json:"errors,omitempty"`
	Metadata     TransformMetadata   `json:"metadata"`
}

type TransformError struct {
	TransactionType string `json:"transactionType"`
	Error           string `json:"error"`
	Code            string `json:"code"`
}

type AcuanRuleConfig struct {
	TransformationType string                   `json:"transformationType"`
	Transactions       []AcuanTransactionConfig `json:"transactions"`
	Validations        []AcuanValidationConfig  `json:"validations"`
}

type AcuanTransactionConfig struct {
	FromAccount     AcuanAccountConfig     `json:"fromAccount"`
	ToAccount       AcuanAccountConfig     `json:"toAccount"`
	Amount          AcuanAmountConfig      `json:"amount"`
	TransactionType string                 `json:"transactionType"`
	OrderType       string                 `json:"orderType"`
	Description     AcuanDescriptionConfig `json:"description"`
	Metadata        AcuanMetadataConfig    `json:"metadata"`
}

type AcuanAccountConfig struct {
	Type    string                    `json:"type"`
	Path    string                    `json:"path,omitempty"`
	Source  string                    `json:"source,omitempty"`
	Mapping AcuanAccountMappingConfig `json:"mapping,omitempty"`
}

type AcuanAccountMappingConfig struct {
	ConfigPath   string `json:"configPath"`
	EntitySource string `json:"entitySource,omitempty"`
}

type AcuanAmountConfig struct {
	Type   string `json:"type"`
	Source string `json:"source,omitempty"`
	Value  string `json:"value,omitempty"`
}

type AcuanDescriptionConfig struct {
	Type     string `json:"type"`
	Value    string `json:"value,omitempty"`
	Template string `json:"template,omitempty"`
	Source   string `json:"source,omitempty"`
}

type AcuanMetadataConfig struct {
	Type    string                 `json:"type"`
	Sources []interface{}          `json:"sources,omitempty"`
	Value   map[string]interface{} `json:"value,omitempty"`
}

type AcuanValidationConfig struct {
	Type      string `json:"type"`
	Field     string `json:"field"`
	ErrorCode string `json:"errorCode"`
}

type AcuanRule struct {
	ID              string
	TransactionType string
	RuleName        string
	Description     string
	Version         int
	IsActive        bool
	Config          json.RawMessage
	Tags            json.RawMessage
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CreatedBy       string
}

type AcuanRuleVersion struct {
	ID          string
	RuleID      string
	Version     int
	Config      json.RawMessage
	CreatedAt   time.Time
	CreatedBy   string
	ChangeNotes string
}

type CreateAcuanRuleRequest struct {
	TransactionType string
	RuleName        string
	Description     string
	Config          json.RawMessage
	Tags            json.RawMessage
	CreatedBy       string
}

type UpdateAcuanRuleRequest struct {
	Config      json.RawMessage
	ChangeNotes string
	UpdatedBy   string
}

type ListAcuanRulesRequest struct {
	TransactionTypes []string
	IsActive         *bool
	Tags             map[string]interface{}
	Limit            int
	Offset           int
}
