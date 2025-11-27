package grule

import (
	"context"
	"fmt"

	"bitbucket.org/Amartha/go-megatron/internal/domain/transformation"

	"github.com/google/uuid"
	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/pkg"
)

type GruleTransformationEngine struct {
	lib *ast.KnowledgeLibrary
}

func NewGruleTransformationEngine() *GruleTransformationEngine {
	return &GruleTransformationEngine{
		lib: ast.NewKnowledgeLibrary(),
	}
}

func (g *GruleTransformationEngine) Transform(
	ctx context.Context,
	wt *transformation.WalletTransaction,
	ruleContent string,
) ([]*transformation.Transaction, error) {

	rb := builder.NewRuleBuilder(g.lib)
	resource := pkg.NewBytesResource([]byte(ruleContent))

	ruleName := wt.TransactionType().String()

	if err := rb.BuildRuleFromResource(ruleName, "1.0.0", resource); err != nil {
		return nil, fmt.Errorf("failed to build rule: %w", err)
	}

	kb, err := g.lib.NewKnowledgeBaseInstance(ruleName, "1.0.0")
	if err != nil {
		return nil, fmt.Errorf("failed to create knowledge base: %w", err)
	}

	dataCtx, journal := g.prepareDataContext(wt)

	eng := engine.NewGruleEngine()
	if err := eng.Execute(dataCtx, kb); err != nil {
		return nil, fmt.Errorf("failed to execute rule: %w", err)
	}

	return g.convertToTransactions(journal, wt)
}

func (g *GruleTransformationEngine) prepareDataContext(
	wt *transformation.WalletTransaction,
) (ast.IDataContext, *GruleJournal) {

	acuan := &GruleAcuan{
		ID:                       wt.ID().String(),
		Status:                   wt.Status().String(),
		AccountNumber:            wt.AccountNumber().String(),
		SourceAccountId:          wt.AccountNumber().String(),
		DestinationAccountNumber: wt.DestinationAccountNumber().String(),
		RefNumber:                wt.RefNumber().String(),
		TransactionType:          wt.TransactionType().String(),
		TransactionTime:          wt.TransactionTime(),
		TransactionFlow:          wt.TransactionFlow().String(),
		NetAmount:                wt.NetAmount().Value(),
		Amount:                   wt.NetAmount().Value(),
		Currency:                 wt.NetAmount().Currency().Code(),
		Description:              wt.Description(),
		Metadata:                 wt.Metadata().ToMap(),
		CreatedAt:                wt.CreatedAt(),
	}

	journal := &GruleJournal{
		Transactions: &GruleTransactionList{
			Items: []GruleJournalEntry{},
		},
	}

	transaction := &GruleTransaction{
		IsReadyToPublish: false,
	}

	journalDebit1 := &GruleJournalDebitCredit{
		Amount:          wt.NetAmount().Value(),
		TransactionDate: wt.TransactionTime().Format("2006-01-02"),
		Status:          wt.Status().ToGruleFormat(),
		TypeTransaction: wt.TransactionType().String(),
		RefNumber:       wt.RefNumber().String(),
		Description:     wt.Description(),
		Metadata:        wt.Metadata().ToMap(),
		OrderTime:       wt.CreatedAt(),
		TransactionTime: wt.TransactionTime(),
		Currency:        wt.NetAmount().Currency().Code(),
		TransactionID:   uuid.New().String(),
	}

	journalCredit1 := &GruleJournalDebitCredit{
		Amount:          wt.NetAmount().Value(),
		TransactionDate: wt.TransactionTime().Format("2006-01-02"),
		Status:          wt.Status().ToGruleFormat(),
		TypeTransaction: wt.TransactionType().String(),
		RefNumber:       wt.RefNumber().String(),
		Description:     wt.Description(),
		Metadata:        wt.Metadata().ToMap(),
		OrderTime:       wt.CreatedAt(),
		TransactionTime: wt.TransactionTime(),
		Currency:        wt.NetAmount().Currency().Code(),
		TransactionID:   uuid.New().String(),
	}

	journalDebit2 := &GruleJournalDebitCredit{
		Amount:          wt.NetAmount().Value(),
		TransactionDate: wt.TransactionTime().Format("2006-01-02"),
		Status:          wt.Status().ToGruleFormat(),
		TypeTransaction: wt.TransactionType().String(),
		RefNumber:       wt.RefNumber().String(),
		Description:     wt.Description(),
		Metadata:        wt.Metadata().ToMap(),
		OrderTime:       wt.CreatedAt(),
		TransactionTime: wt.TransactionTime(),
		Currency:        wt.NetAmount().Currency().Code(),
		TransactionID:   uuid.New().String(),
	}

	journalCredit2 := &GruleJournalDebitCredit{
		Amount:          wt.NetAmount().Value(),
		TransactionDate: wt.TransactionTime().Format("2006-01-02"),
		Status:          wt.Status().ToGruleFormat(),
		TypeTransaction: wt.TransactionType().String(),
		RefNumber:       wt.RefNumber().String(),
		Description:     wt.Description(),
		Metadata:        wt.Metadata().ToMap(),
		OrderTime:       wt.CreatedAt(),
		TransactionTime: wt.TransactionTime(),
		Currency:        wt.NetAmount().Currency().Code(),
		TransactionID:   uuid.New().String(),
	}

	dataCtx := ast.NewDataContext()
	dataCtx.Add("Acuan", acuan)
	dataCtx.Add("Journal", journal)
	dataCtx.Add("Transaction", transaction)
	dataCtx.Add("JournalDebit1", journalDebit1)
	dataCtx.Add("JournalCredit1", journalCredit1)
	dataCtx.Add("JournalDebit2", journalDebit2)
	dataCtx.Add("JournalCredit2", journalCredit2)

	return dataCtx, journal
}

func (g *GruleTransformationEngine) convertToTransactions(
	journal *GruleJournal,
	wt *transformation.WalletTransaction,
) ([]*transformation.Transaction, error) {

	var transactions []*transformation.Transaction

	for _, entry := range journal.Transactions.Items {
		tx, err := transformation.NewTransaction(
			entry.TransactionID,
			entry.Account,
			entry.Account,
			wt.TransactionTime(),
			entry.Amount,
			entry.Currency,
			entry.Status,
			entry.TypeTransaction,
			entry.Description,
			entry.RefNumber,
			entry.OrderType,
			entry.OrderTime,
			entry.TransactionTime,
			entry.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create transaction: %w", err)
		}

		if entry.Narrative != "" {
			tx.SetNarratives(entry.Narrative, entry.Narrative)
		}

		transactions = append(transactions, tx)
	}

	return transactions, nil
}
