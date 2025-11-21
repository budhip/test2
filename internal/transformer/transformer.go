package transformer

import (
	"bitbucket.org/Amartha/go-megatron/internal/acuanrepository"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	models "bitbucket.org/Amartha/go-megatron/internal/megatron"
)

// Engine adalah rule engine untuk execute transformation
type Engine struct {
	ruleRepo acuanrepository.RuleRepository
	config   Config
}

type Config struct {
	DefaultCurrency string
	Timeout         time.Duration
}

func NewEngine(ruleRepo acuanrepository.RuleRepository, config Config) *Engine {
	return &Engine{
		ruleRepo: ruleRepo,
		config:   config,
	}
}

// Transform executes transformation based on rule
func (e *Engine) Transform(ctx context.Context, req models.TransformRequest) (*models.TransformResponse, error) {
	startTime := time.Now()

	// Get active rule untuk transaction type
	rule, err := e.ruleRepo.GetActiveRule(ctx, req.TransactionType)
	if err != nil {
		return nil, fmt.Errorf("failed to get rule for %s: %w", req.TransactionType, err)
	}

	// Parse rule config
	var config RuleConfig
	if err := json.Unmarshal(rule.Config, &config); err != nil {
		return nil, fmt.Errorf("failed to parse rule config: %w", err)
	}

	// Validate input
	if err := e.validate(req, config); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Execute transformation
	transactions, err := e.executeTransformation(ctx, req, config)
	if err != nil {
		return nil, fmt.Errorf("transformation failed: %w", err)
	}

	executionTime := time.Since(startTime)

	return &models.TransformResponse{
		Transactions: transactions,
		Metadata: models.TransformMetadata{
			RuleID:          rule.ID,
			RuleName:        rule.RuleName,
			RuleVersion:     rule.Version,
			ExecutionTimeMs: int(executionTime.Milliseconds()),
		},
	}, nil
}

// executeTransformation menjalankan actual transformation
func (e *Engine) executeTransformation(ctx context.Context, req models.TransformRequest, config RuleConfig) ([]models.TransactionOutput, error) {
	var results []models.TransactionOutput

	for _, txConfig := range config.Transactions {
		tx, err := e.buildTransaction(req, txConfig)
		if err != nil {
			return nil, err
		}
		results = append(results, tx)
	}

	return results, nil
}

// buildTransaction membuat transaction berdasarkan config
func (e *Engine) buildTransaction(req models.TransformRequest, txConfig TransactionConfig) (models.TransactionOutput, error) {
	tx := models.TransactionOutput{
		TransactionID:   uuid.New().String(),
		TransactionTime: req.ParentTransaction.TransactionTime,
		OrderTime:       time.Now(),
		Currency:        e.resolveCurrency(req.Amount.Currency),
		Status:          req.ParentTransaction.Status,
		RefNumber:       req.ParentTransaction.RefNumber,
	}

	// Resolve FromAccount
	fromAccount, err := e.resolveAccount(req, txConfig.FromAccount)
	if err != nil {
		return tx, fmt.Errorf("failed to resolve fromAccount: %w", err)
	}
	tx.FromAccount = fromAccount

	// Resolve ToAccount
	toAccount, err := e.resolveAccount(req, txConfig.ToAccount)
	if err != nil {
		return tx, fmt.Errorf("failed to resolve toAccount: %w", err)
	}
	tx.ToAccount = toAccount

	// Resolve Amount
	amount, err := e.resolveAmount(req, txConfig.Amount)
	if err != nil {
		return tx, fmt.Errorf("failed to resolve amount: %w", err)
	}
	tx.Amount = amount

	// Set static fields
	tx.TypeTransaction = txConfig.TransactionType
	tx.OrderType = txConfig.OrderType

	// Resolve Description
	description, err := e.resolveDescription(req, txConfig.Description)
	if err != nil {
		return tx, fmt.Errorf("failed to resolve description: %w", err)
	}
	tx.Description = description

	// Resolve TransactionDate
	tx.TransactionDate = req.ParentTransaction.TransactionTime.Format("2006-01-02")

	// Resolve Metadata
	metadata, err := e.resolveMetadata(req, txConfig.Metadata)
	if err != nil {
		return tx, fmt.Errorf("failed to resolve metadata: %w", err)
	}
	tx.Metadata = metadata

	return tx, nil
}

// resolveAccount resolves account berdasarkan config
func (e *Engine) resolveAccount(req models.TransformRequest, config AccountConfig) (string, error) {
	switch config.Type {
	case "config":
		// Get from context config
		return e.getFromConfig(req.Context, config.Path)

	case "dynamic":
		if config.Source == "entity" {
			// Get entity from parent transaction
			entity := req.ParentTransaction.Account.Entity
			if entity == "" {
				return "", fmt.Errorf("entity not found in parent transaction account")
			}

			// Map entity to actual entity code
			if config.Mapping.EntitySource != "" {
				mappedEntity, err := e.resolveValue(req, config.Mapping.EntitySource)
				if err == nil && mappedEntity != "" {
					entity = mappedEntity
				}
			}

			// Get account number from mapping
			configPath := config.Mapping.ConfigPath
			mapping, err := e.getConfigMapping(req.Context, configPath)
			if err != nil {
				return "", err
			}

			accountNumber, ok := mapping[entity]
			if !ok {
				return "", fmt.Errorf("account number not found for entity %s in %s", entity, configPath)
			}
			return accountNumber, nil
		}

		// Resolve from other sources
		return e.resolveValue(req, config.Source)

	case "input":
		// Get from parent transaction using path
		// ← PERBAIKAN DI SINI: Gunakan resolveValue untuk handle semua input paths
		return e.resolveValue(req, config.Source)

	default:
		return "", fmt.Errorf("unsupported account type: %s", config.Type)
	}
}

// resolveAmount resolves amount berdasarkan config
func (e *Engine) resolveAmount(req models.TransformRequest, config AmountConfig) (decimal.Decimal, error) {
	switch config.Type {
	case "input":
		// Use amount dari request
		return req.Amount.Value, nil

	case "static":
		// Parse static value
		return decimal.NewFromString(config.Value)

	case "calculate":
		// TODO: Implement calculation logic
		return decimal.Zero, fmt.Errorf("calculation not yet implemented")

	default:
		return decimal.Zero, fmt.Errorf("unsupported amount type: %s", config.Type)
	}
}

// resolveDescription resolves description berdasarkan config
func (e *Engine) resolveDescription(req models.TransformRequest, config DescriptionConfig) (string, error) {
	switch config.Type {
	case "static":
		return config.Value, nil

	case "template":
		// Simple template replacement
		result := config.Template
		result = strings.ReplaceAll(result, "{{parentTransaction.accountNumber}}", req.ParentTransaction.AccountNumber)
		result = strings.ReplaceAll(result, "{{parentTransaction.destinationAccountNumber}}", req.ParentTransaction.DestinationAccountNumber)
		result = strings.ReplaceAll(result, "{{parentTransaction.refNumber}}", req.ParentTransaction.RefNumber)
		result = strings.ReplaceAll(result, "{{parentTransaction.description}}", req.ParentTransaction.Description)
		result = strings.ReplaceAll(result, "{{parentTransaction.transactionType}}", req.ParentTransaction.TransactionType)

		// Handle metadata templates: {{parentTransaction.metadata.loan_account_number}}
		for key, value := range req.ParentTransaction.Metadata {
			placeholder := fmt.Sprintf("{{parentTransaction.metadata.%s}}", key)
			result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
		}

		return result, nil

	case "input":
		// Get description from input path
		if config.Source == "" {
			// Default to parentTransaction.description if no source specified
			return req.ParentTransaction.Description, nil
		}
		return e.resolveValue(req, config.Source)

	default:
		return "", fmt.Errorf("unsupported description type: %s", config.Type)
	}
}

// resolveMetadata resolves metadata berdasarkan config
func (e *Engine) resolveMetadata(req models.TransformRequest, config MetadataConfig) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	switch config.Type {
	case "merge":
		// Merge dari berbagai sources
		for _, source := range config.Sources {
			if sourceStr, ok := source.(string); ok {
				// It's a path reference
				if sourceStr == "parentTransaction.metadata" {
					for k, v := range req.ParentTransaction.Metadata {
						result[k] = v
					}
				} else {
					// Try to resolve as path
					value, err := e.resolveValue(req, sourceStr)
					if err == nil && value != "" {
						// Extract key name from path (e.g., "loan_account_number" from path)
						parts := strings.Split(sourceStr, ".")
						key := parts[len(parts)-1]
						result[key] = value
					}
				}
			} else if sourceMap, ok := source.(map[string]interface{}); ok {
				// It's a dynamic config
				if sourceMap["type"] == "dynamic" {
					field := sourceMap["field"].(string)
					sourcePath := sourceMap["source"].(string)
					value, err := e.resolveValue(req, sourcePath)
					if err == nil && value != "" {
						result[field] = value
					}
				}
			}
		}

	case "static":
		result = config.Value

	default:
		return nil, fmt.Errorf("unsupported metadata type: %s", config.Type)
	}

	return result, nil
}

// Helper functions

func (e *Engine) resolveValue(req models.TransformRequest, path string) (string, error) {
	parts := strings.Split(path, ".")

	if parts[0] == "parentTransaction" {
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid path: %s", path)
		}

		switch parts[1] {
		case "id":
			return req.ParentTransaction.ID, nil
		case "accountNumber":
			return req.ParentTransaction.AccountNumber, nil
		case "destinationAccountNumber":
			return req.ParentTransaction.DestinationAccountNumber, nil
		case "refNumber":
			return req.ParentTransaction.RefNumber, nil
		case "description":
			return req.ParentTransaction.Description, nil
		case "transactionType":
			return req.ParentTransaction.TransactionType, nil
		case "transactionFlow":
			return req.ParentTransaction.TransactionFlow, nil
		case "status":
			return req.ParentTransaction.Status, nil
		case "account":
			if len(parts) < 3 {
				return "", fmt.Errorf("invalid account path: %s", path)
			}
			switch parts[2] {
			case "accountNumber":
				return req.ParentTransaction.Account.AccountNumber, nil
			case "name":
				return req.ParentTransaction.Account.Name, nil
			case "entity":
				return req.ParentTransaction.Account.Entity, nil
			case "categoryCode":
				return req.ParentTransaction.Account.CategoryCode, nil
			case "subCategoryCode":
				return req.ParentTransaction.Account.SubCategoryCode, nil
			default:
				return "", fmt.Errorf("unsupported account field: %s", parts[2])
			}
		case "metadata":
			// Handle metadata access: parentTransaction.metadata.loan_account_number
			if len(parts) < 3 {
				return "", fmt.Errorf("invalid metadata path: %s", path)
			}
			metadataKey := parts[2]
			if value, ok := req.ParentTransaction.Metadata[metadataKey]; ok {
				// Convert to string
				switch v := value.(type) {
				case string:
					return v, nil
				case float64:
					return fmt.Sprintf("%v", v), nil
				case int:
					return fmt.Sprintf("%d", v), nil
				case bool:
					return fmt.Sprintf("%t", v), nil
				default:
					return fmt.Sprintf("%v", v), nil
				}
			}
			return "", fmt.Errorf("metadata key not found: %s", metadataKey)
		default:
			return "", fmt.Errorf("unsupported parentTransaction field: %s", parts[1])
		}
	}

	if parts[0] == "amount" {
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid amount path: %s", path)
		}
		switch parts[1] {
		case "value":
			return req.Amount.Value.String(), nil
		case "currency":
			return req.Amount.Currency, nil
		default:
			return "", fmt.Errorf("unsupported amount field: %s", parts[1])
		}
	}

	return "", fmt.Errorf("unsupported path: %s", path)
}

func (e *Engine) getFromConfig(ctx models.TransformContext, path string) (string, error) {
	switch path {
	case "AccountConfig.SystemAccountNumber":
		return ctx.SystemAccountNumber, nil
	default:
		return "", fmt.Errorf("unsupported config path: %s", path)
	}
}

func (e *Engine) getConfigMapping(ctx models.TransformContext, path string) (map[string]string, error) {
	switch path {
	case "AccountConfig.AccountNumberInsurancePremiumDisbursementByEntity":
		return ctx.AccountNumberInsurancePremiumDisbursementByEntity, nil
	case "AccountConfig.MapAccountEntity":
		return ctx.MapAccountEntity, nil
	default:
		return nil, fmt.Errorf("unsupported config mapping path: %s", path)
	}
}

func (e *Engine) validate(req models.TransformRequest, config RuleConfig) error {
	for _, validation := range config.Validations {
		switch validation.Type {
		case "required":
			value, err := e.resolveValue(req, validation.Field)
			if err != nil || value == "" {
				return fmt.Errorf("validation failed: %s is required (error code: %s)",
					validation.Field, validation.ErrorCode)
			}
		}
	}
	return nil
}

func (e *Engine) resolveCurrency(currency string) string {
	if currency == "" {
		return e.config.DefaultCurrency
	}
	return currency
}

// RuleConfig structures

type RuleConfig struct {
	TransformationType string              `json:"transformationType"`
	Transactions       []TransactionConfig `json:"transactions"`
	Validations        []ValidationConfig  `json:"validations"`
}

type TransactionConfig struct {
	FromAccount     AccountConfig     `json:"fromAccount"`
	ToAccount       AccountConfig     `json:"toAccount"`
	Amount          AmountConfig      `json:"amount"`
	TransactionType string            `json:"transactionType"`
	OrderType       string            `json:"orderType"`
	Description     DescriptionConfig `json:"description"`
	Metadata        MetadataConfig    `json:"metadata"`
}

type AccountConfig struct {
	Type    string               `json:"type"` // "config", "dynamic", "input"
	Path    string               `json:"path,omitempty"`
	Source  string               `json:"source,omitempty"`
	Mapping AccountMappingConfig `json:"mapping,omitempty"`
}

type AccountMappingConfig struct {
	ConfigPath   string `json:"configPath"`
	EntitySource string `json:"entitySource,omitempty"`
}

type AmountConfig struct {
	Type   string `json:"type"` // "input", "static", "calculate"
	Source string `json:"source,omitempty"`
	Value  string `json:"value,omitempty"`
}

type DescriptionConfig struct {
	Type     string `json:"type"` // "static", "template"
	Value    string `json:"value,omitempty"`
	Template string `json:"template,omitempty"`
	Source   string `json:"source,omitempty"`
}

type MetadataConfig struct {
	Type    string                 `json:"type"` // "merge", "static"
	Sources []interface{}          `json:"sources,omitempty"`
	Value   map[string]interface{} `json:"value,omitempty"`
}

type ValidationConfig struct {
	Type      string `json:"type"` // "required", "pattern", "range"
	Field     string `json:"field"`
	ErrorCode string `json:"errorCode"`
}
