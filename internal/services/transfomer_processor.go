package services

import (
	"bitbucket.org/Amartha/go-megatron/internal/repositories"

	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	models "bitbucket.org/Amartha/go-megatron/internal/models"
)

type Engine struct {
	ruleRepo repositories.AcuanRuleRepository
	config   Config
}

type Config struct {
	DefaultCurrency string
	Timeout         time.Duration
}

func NewEngine(ruleRepo repositories.AcuanRuleRepository, config Config) *Engine {
	return &Engine{
		ruleRepo: ruleRepo,
		config:   config,
	}
}

// TransformWalletTransaction transforms wallet transaction to acuan transactions
func (e *Engine) TransformWalletTransaction(ctx context.Context, req models.WalletTransactionRequest) (*models.WalletTransactionResponse, error) {
	startTime := time.Now()

	var allTransactions []models.TransactionReq

	// Transform NetAmount
	if !req.WalletTransaction.NetAmount.Value.IsZero() {
		rule, err := e.ruleRepo.GetActiveRule(ctx, req.WalletTransaction.TransactionType)
		if err != nil {
			return nil, fmt.Errorf("failed to get rule for %s: %w", req.WalletTransaction.TransactionType, err)
		}

		var config models.AcuanRuleConfig
		if err := json.Unmarshal(rule.Config, &config); err != nil {
			return nil, fmt.Errorf("failed to parse rule config: %w", err)
		}

		transactions, err := e.executeWalletTransformation(ctx, req, config, req.WalletTransaction.NetAmount.Value, req.WalletTransaction.TransactionType)
		if err != nil {
			return nil, fmt.Errorf("transformation failed for net amount: %w", err)
		}
		allTransactions = append(allTransactions, transactions...)
	}

	// Transform Amounts array
	for _, amountItem := range req.WalletTransaction.Amounts {
		if amountItem.Amount.Value.IsZero() {
			continue
		}

		rule, err := e.ruleRepo.GetActiveRule(ctx, amountItem.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to get rule for %s: %w", amountItem.Type, err)
		}

		var config models.AcuanRuleConfig
		if err := json.Unmarshal(rule.Config, &config); err != nil {
			return nil, fmt.Errorf("failed to parse rule config: %w", err)
		}

		transactions, err := e.executeWalletTransformation(ctx, req, config, amountItem.Amount.Value, amountItem.Type)
		if err != nil {
			return nil, fmt.Errorf("transformation failed for %s: %w", amountItem.Type, err)
		}
		allTransactions = append(allTransactions, transactions...)
	}

	executionTime := time.Since(startTime)

	return &models.WalletTransactionResponse{
		Transactions: allTransactions,
		Metadata: models.TransformMetadata{
			ExecutionTimeMs: int(executionTime.Milliseconds()),
		},
	}, nil
}

func (e *Engine) executeWalletTransformation(ctx context.Context, req models.WalletTransactionRequest, config models.AcuanRuleConfig, amount decimal.Decimal, transactionType string) ([]models.TransactionReq, error) {
	var results []models.TransactionReq

	for _, txConfig := range config.Transactions {
		tx, err := e.buildWalletTransaction(req, txConfig, amount, transactionType)
		if err != nil {
			return nil, err
		}
		results = append(results, tx)
	}

	return results, nil
}

func (e *Engine) buildWalletTransaction(req models.WalletTransactionRequest, txConfig models.AcuanTransactionConfig, amount decimal.Decimal, transactionType string) (models.TransactionReq, error) {
	tx := models.TransactionReq{
		TransactionID:   uuid.New().String(),
		TransactionTime: req.WalletTransaction.TransactionTime,
		OrderTime:       req.WalletTransaction.CreatedAt,
		Currency:        e.resolveCurrency(req.WalletTransaction.NetAmount.Currency),
		Status:          e.transformStatus(req.WalletTransaction.Status),
		RefNumber:       req.WalletTransaction.RefNumber,
		Description:     req.WalletTransaction.Description,
		Metadata:        req.WalletTransaction.Metadata,
	}

	// Resolve FromAccount
	fromAccount, err := e.resolveWalletAccount(req, txConfig.FromAccount)
	if err != nil {
		return tx, fmt.Errorf("failed to resolve fromAccount: %w", err)
	}
	tx.FromAccount = fromAccount

	// Resolve ToAccount
	toAccount, err := e.resolveWalletAccount(req, txConfig.ToAccount)
	if err != nil {
		return tx, fmt.Errorf("failed to resolve toAccount: %w", err)
	}
	tx.ToAccount = toAccount

	// Set Amount
	tx.Amount = amount

	// Set transaction type and order type
	tx.TypeTransaction = transactionType
	tx.OrderType = txConfig.OrderType

	// Resolve TransactionDate
	tx.TransactionDate = req.WalletTransaction.TransactionTime.Format("2006-01-02")

	// Add entity to metadata if available
	if entity, ok := req.WalletTransaction.Metadata["entity"]; ok {
		if tx.Metadata == nil {
			tx.Metadata = make(map[string]interface{})
		}
		tx.Metadata["entity"] = entity
	}

	return tx, nil
}

func (e *Engine) resolveWalletAccount(req models.WalletTransactionRequest, config models.AcuanAccountConfig) (string, error) {
	switch config.Type {
	case "config":
		return e.getWalletConfig(req.Context, config.Path)
	case "input":
		return e.resolveWalletValue(req, config.Source)
	default:
		return "", fmt.Errorf("unsupported account type: %s", config.Type)
	}
}

func (e *Engine) resolveWalletValue(req models.WalletTransactionRequest, path string) (string, error) {
	parts := strings.Split(path, ".")

	if parts[0] == "walletTransaction" {
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid path: %s", path)
		}

		switch parts[1] {
		case "id":
			return req.WalletTransaction.ID, nil
		case "accountNumber":
			return req.WalletTransaction.AccountNumber, nil
		case "destinationAccountNumber":
			return req.WalletTransaction.DestinationAccountNumber, nil
		case "refNumber":
			return req.WalletTransaction.RefNumber, nil
		case "description":
			return req.WalletTransaction.Description, nil
		case "transactionType":
			return req.WalletTransaction.TransactionType, nil
		case "transactionFlow":
			return req.WalletTransaction.TransactionFlow, nil
		case "status":
			return req.WalletTransaction.Status, nil
		case "metadata":
			if len(parts) < 3 {
				return "", fmt.Errorf("invalid metadata path: %s", path)
			}
			metadataKey := parts[2]
			if value, ok := req.WalletTransaction.Metadata[metadataKey]; ok {
				return fmt.Sprintf("%v", value), nil
			}
			return "", fmt.Errorf("metadata key not found: %s", metadataKey)
		default:
			return "", fmt.Errorf("unsupported walletTransaction field: %s", parts[1])
		}
	}

	return "", fmt.Errorf("unsupported path: %s", path)
}

func (e *Engine) getWalletConfig(ctx models.TransformContext, path string) (string, error) {
	switch path {
	case "AccountConfig.SystemAccountNumber":
		return ctx.SystemAccountNumber, nil
	default:
		return "", fmt.Errorf("unsupported config path: %s", path)
	}
}

func (e *Engine) transformStatus(status string) string {
	if status == "SUCCESS" {
		return "1"
	}
	return "0"
}

// Transform executes transformation based on rule
func (e *Engine) Transform(ctx context.Context, req models.TransformRequest) (*models.TransformResponse, error) {
	startTime := time.Now()

	rule, err := e.ruleRepo.GetActiveRule(ctx, req.TransactionType)
	if err != nil {
		return nil, fmt.Errorf("failed to get rule for %s: %w", req.TransactionType, err)
	}

	var config models.AcuanRuleConfig
	if err := json.Unmarshal(rule.Config, &config); err != nil {
		return nil, fmt.Errorf("failed to parse rule config: %w", err)
	}

	if err := e.validate(req, config); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

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

func (e *Engine) executeTransformation(ctx context.Context, req models.TransformRequest, config models.AcuanRuleConfig) ([]models.TransactionOutput, error) {
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

func (e *Engine) buildTransaction(req models.TransformRequest, txConfig models.AcuanTransactionConfig) (models.TransactionOutput, error) {
	tx := models.TransactionOutput{
		TransactionID:   uuid.New().String(),
		TransactionTime: req.ParentTransaction.TransactionTime,
		OrderTime:       time.Now(),
		Currency:        e.resolveCurrency(req.Amount.Currency),
		Status:          req.ParentTransaction.Status,
		RefNumber:       req.ParentTransaction.RefNumber,
	}

	fromAccount, err := e.resolveAccount(req, txConfig.FromAccount)
	if err != nil {
		return tx, fmt.Errorf("failed to resolve fromAccount: %w", err)
	}
	tx.FromAccount = fromAccount

	toAccount, err := e.resolveAccount(req, txConfig.ToAccount)
	if err != nil {
		return tx, fmt.Errorf("failed to resolve toAccount: %w", err)
	}
	tx.ToAccount = toAccount

	amount, err := e.resolveAmount(req, txConfig.Amount)
	if err != nil {
		return tx, fmt.Errorf("failed to resolve amount: %w", err)
	}
	tx.Amount = amount

	tx.TypeTransaction = txConfig.TransactionType
	tx.OrderType = txConfig.OrderType

	description, err := e.resolveDescription(req, txConfig.Description)
	if err != nil {
		return tx, fmt.Errorf("failed to resolve description: %w", err)
	}
	tx.Description = description

	tx.TransactionDate = req.ParentTransaction.TransactionTime.Format("2006-01-02")

	metadata, err := e.resolveMetadata(req, txConfig.Metadata)
	if err != nil {
		return tx, fmt.Errorf("failed to resolve metadata: %w", err)
	}
	tx.Metadata = metadata

	return tx, nil
}

func (e *Engine) resolveAccount(req models.TransformRequest, config models.AcuanAccountConfig) (string, error) {
	switch config.Type {
	case "config":
		return e.getFromConfig(req.Context, config.Path)

	case "dynamic":
		if config.Source == "entity" {
			entity := req.ParentTransaction.Account.Entity
			if entity == "" {
				return "", fmt.Errorf("entity not found in parent transaction account")
			}

			if config.Mapping.EntitySource != "" {
				mappedEntity, err := e.resolveValue(req, config.Mapping.EntitySource)
				if err == nil && mappedEntity != "" {
					entity = mappedEntity
				}
			}

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

		return e.resolveValue(req, config.Source)

	case "input":
		return e.resolveValue(req, config.Source)

	default:
		return "", fmt.Errorf("unsupported account type: %s", config.Type)
	}
}

func (e *Engine) resolveAmount(req models.TransformRequest, config models.AcuanAmountConfig) (decimal.Decimal, error) {
	switch config.Type {
	case "input":
		return req.Amount.Value, nil

	case "static":
		return decimal.NewFromString(config.Value)

	case "calculate":
		return decimal.Zero, fmt.Errorf("calculation not yet implemented")

	default:
		return decimal.Zero, fmt.Errorf("unsupported amount type: %s", config.Type)
	}
}

func (e *Engine) resolveDescription(req models.TransformRequest, config models.AcuanDescriptionConfig) (string, error) {
	switch config.Type {
	case "static":
		return config.Value, nil

	case "template":
		result := config.Template
		result = strings.ReplaceAll(result, "{{parentTransaction.accountNumber}}", req.ParentTransaction.AccountNumber)
		result = strings.ReplaceAll(result, "{{parentTransaction.destinationAccountNumber}}", req.ParentTransaction.DestinationAccountNumber)
		result = strings.ReplaceAll(result, "{{parentTransaction.refNumber}}", req.ParentTransaction.RefNumber)
		result = strings.ReplaceAll(result, "{{parentTransaction.description}}", req.ParentTransaction.Description)
		result = strings.ReplaceAll(result, "{{parentTransaction.transactionType}}", req.ParentTransaction.TransactionType)

		for key, value := range req.ParentTransaction.Metadata {
			placeholder := fmt.Sprintf("{{parentTransaction.metadata.%s}}", key)
			result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
		}

		return result, nil

	case "input":
		if config.Source == "" {
			return req.ParentTransaction.Description, nil
		}
		return e.resolveValue(req, config.Source)

	default:
		return "", fmt.Errorf("unsupported description type: %s", config.Type)
	}
}

func (e *Engine) resolveMetadata(req models.TransformRequest, config models.AcuanMetadataConfig) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	switch config.Type {
	case "merge":
		for _, source := range config.Sources {
			if sourceStr, ok := source.(string); ok {
				if sourceStr == "parentTransaction.metadata" {
					for k, v := range req.ParentTransaction.Metadata {
						result[k] = v
					}
				} else {
					parts := strings.Split(sourceStr, ".")
					key := parts[len(parts)-1]
					value, err := e.resolveValue(req, sourceStr)
					if err == nil && value != "" {
						result[key] = value
					}
				}
			} else if sourceMap, ok := source.(map[string]interface{}); ok {
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
			if len(parts) < 3 {
				return "", fmt.Errorf("invalid metadata path: %s", path)
			}
			metadataKey := parts[2]
			if value, ok := req.ParentTransaction.Metadata[metadataKey]; ok {
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

func (e *Engine) validate(req models.TransformRequest, config models.AcuanRuleConfig) error {
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
