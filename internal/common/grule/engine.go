package grule

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"time"

	"bitbucket.org/Amartha/go-megatron/internal/models"
	"bitbucket.org/Amartha/go-megatron/internal/repositories"

	"github.com/google/uuid"
	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/pkg"
)

type Engine struct {
	ruleRepo repositories.AcuanRuleRepository
}

func NewEngine(ruleRepo repositories.AcuanRuleRepository) *Engine {
	return &Engine{
		ruleRepo: ruleRepo,
	}
}

func (e *Engine) TransformWalletTransaction(ctx context.Context, req models.WalletTransactionRequest) (*models.WalletTransactionResponse, error) {
	startTime := time.Now()

	var allTransactions []models.TransactionReq

	// Transform NetAmount first
	if !req.WalletTransaction.NetAmount.Value.IsZero() {
		txs, err := e.executeTransformation(ctx, req, req.WalletTransaction.NetAmount.Value, req.WalletTransaction.TransactionType)
		if err != nil {
			return nil, fmt.Errorf("failed to transform net amount: %w", err)
		}
		allTransactions = append(allTransactions, txs...)
	}

	// Transform each Amount in Amounts array
	for _, amountItem := range req.WalletTransaction.Amounts {
		if amountItem.Amount.Value.IsZero() {
			continue
		}

		txs, err := e.executeTransformation(ctx, req, amountItem.Amount.Value, amountItem.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to transform %s: %w", amountItem.Type, err)
		}
		allTransactions = append(allTransactions, txs...)
	}

	executionTime := time.Since(startTime)

	return &models.WalletTransactionResponse{
		Transactions: allTransactions,
		Metadata: models.TransformMetadata{
			ExecutionTimeMs: int(executionTime.Milliseconds()),
		},
	}, nil
}

func (e *Engine) executeTransformation(ctx context.Context, req models.WalletTransactionRequest, amount decimal.Decimal, transactionType string) ([]models.TransactionReq, error) {
	// Get rule from database
	rule, err := e.ruleRepo.GetActiveRule(ctx, transactionType)
	if err != nil {
		return nil, fmt.Errorf("rule not found for %s: %w", transactionType, err)
	}

	// Parse rule content to get actual Grule content
	var ruleContent string
	if err := json.Unmarshal(rule.Config, &ruleContent); err != nil {
		return nil, fmt.Errorf("failed to parse rule content: %w", err)
	}

	// Create knowledge library
	lib := ast.NewKnowledgeLibrary()
	rb := builder.NewRuleBuilder(lib)

	// Load rule from string
	resource := pkg.NewBytesResource([]byte(ruleContent))
	err = rb.BuildRuleFromResource(transactionType, "1.0.0", resource)
	if err != nil {
		return nil, fmt.Errorf("failed to build rule: %w", err)
	}

	// Create knowledge base
	kb, err := lib.NewKnowledgeBaseInstance(transactionType, "1.0.0")
	if err != nil {
		return nil, fmt.Errorf("failed to create knowledge base: %w", err)
	}

	// Prepare data context
	acuan := &models.GruleAcuan{
		ID:                       req.WalletTransaction.ID,
		Status:                   req.WalletTransaction.Status,
		AccountNumber:            req.WalletTransaction.AccountNumber,
		SourceAccountId:          req.WalletTransaction.AccountNumber,
		DestinationAccountNumber: req.WalletTransaction.DestinationAccountNumber,
		RefNumber:                req.WalletTransaction.RefNumber,
		TransactionType:          transactionType,
		TransactionTime:          req.WalletTransaction.TransactionTime,
		TransactionFlow:          req.WalletTransaction.TransactionFlow,
		NetAmount:                req.WalletTransaction.NetAmount.Value,
		Amount:                   amount,
		Currency:                 req.WalletTransaction.NetAmount.Currency,
		Description:              req.WalletTransaction.Description,
		Metadata:                 req.WalletTransaction.Metadata,
		CreatedAt:                req.WalletTransaction.CreatedAt,
	}

	journal := &models.GruleAcuanJournal{
		Transactions: &models.GruleAcuanTransactionList{
			Items: []models.GruleAcuanJournalEntry{},
		},
	}

	transaction := &models.GruleAcuanTransaction{
		IsReadyToPublish: false,
	}

	// Create helper objects for Grule
	journalDebit1 := e.createJournalEntry(req, amount, transactionType)
	journalCredit1 := e.createJournalEntry(req, amount, transactionType)
	journalDebit2 := e.createJournalEntry(req, amount, transactionType)
	journalCredit2 := e.createJournalEntry(req, amount, transactionType)

	// Create data context
	dataCtx := ast.NewDataContext()
	err = dataCtx.Add("Acuan", acuan)
	if err != nil {
		return nil, err
	}
	err = dataCtx.Add("Journal", journal)
	if err != nil {
		return nil, err
	}
	err = dataCtx.Add("Transaction", transaction)
	if err != nil {
		return nil, err
	}
	err = dataCtx.Add("JournalDebit1", journalDebit1)
	if err != nil {
		return nil, err
	}
	err = dataCtx.Add("JournalCredit1", journalCredit1)
	if err != nil {
		return nil, err
	}
	err = dataCtx.Add("JournalDebit2", journalDebit2)
	if err != nil {
		return nil, err
	}
	err = dataCtx.Add("JournalCredit2", journalCredit2)
	if err != nil {
		return nil, err
	}

	// Execute rule engine
	eng := engine.NewGruleEngine()
	err = eng.Execute(dataCtx, kb)
	if err != nil {
		return nil, fmt.Errorf("failed to execute rule: %w", err)
	}

	// Convert results to TransactionReq
	return e.convertJournalToTransactions(journal, req), nil
}

func (e *Engine) createJournalEntry(req models.WalletTransactionRequest, amount decimal.Decimal, transactionType string) *models.GruleJournalDebitCredit {
	return &models.GruleJournalDebitCredit{
		Amount:          amount,
		TransactionDate: req.WalletTransaction.TransactionTime.Format("2006-01-02"),
		Status:          e.transformStatus(req.WalletTransaction.Status),
		TypeTransaction: transactionType,
		RefNumber:       req.WalletTransaction.RefNumber,
		Description:     req.WalletTransaction.Description,
		Metadata:        req.WalletTransaction.Metadata,
		OrderTime:       req.WalletTransaction.CreatedAt,
		TransactionTime: req.WalletTransaction.TransactionTime,
		Currency:        req.WalletTransaction.NetAmount.Currency,
		TransactionID:   uuid.New().String(),
	}
}

func (e *Engine) convertJournalToTransactions(journal *models.GruleAcuanJournal, req models.WalletTransactionRequest) []models.TransactionReq {
	var results []models.TransactionReq

	for _, entry := range journal.Transactions.Items {
		results = append(results, models.TransactionReq{
			TransactionID:   entry.TransactionID,
			FromAccount:     entry.Account, // Will be set by rule
			ToAccount:       entry.Account, // Will be set by rule
			TransactionDate: entry.TransactionDate,
			Amount:          entry.Amount,
			Status:          entry.Status,
			TypeTransaction: entry.TypeTransaction,
			Description:     entry.Description,
			RefNumber:       entry.RefNumber,
			Metadata:        entry.Metadata,
			OrderTime:       entry.OrderTime,
			OrderType:       entry.OrderType,
			TransactionTime: entry.TransactionTime,
			Currency:        entry.Currency,
		})
	}

	return results
}

func (e *Engine) transformStatus(status string) string {
	if status == "SUCCESS" {
		return "1"
	}
	return "0"
}
