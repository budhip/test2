package grule

import (
	"context"
	"fmt"

	"bitbucket.org/Amartha/go-megatron/internal/domain/transformation"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/pkg"
)

// GruleEngine implements domain.transformation.TransformationEngine
type GruleEngine struct {
	lib *ast.KnowledgeLibrary
}

func NewGruleEngine() *GruleEngine {
	return &GruleEngine{
		lib: ast.NewKnowledgeLibrary(),
	}
}

// Transform implements TransformationEngine interface
func (e *GruleEngine) Transform(
	ctx context.Context,
	walletTx *transformation.WalletTransaction,
	ruleContent string,
) ([]*transformation.Transaction, error) {
	// 1. Build rule from content
	rb := builder.NewRuleBuilder(e.lib)
	resource := pkg.NewBytesResource([]byte(ruleContent))

	ruleName := walletTx.TransactionType().String()

	if err := rb.BuildRuleFromResource(ruleName, "1.0.0", resource); err != nil {
		return nil, fmt.Errorf("failed to build rule: %w", err)
	}

	// 2. Create knowledge base
	kb, err := e.lib.NewKnowledgeBaseInstance(ruleName, "1.0.0")
	if err != nil {
		return nil, err
	}

	// 3. Prepare data context (convert domain entity to Grule format)
	dataCtx, journal := e.prepareDataContext(walletTx)

	// 4. Execute rule engine
	eng := engine.NewGruleEngine()
	if err := eng.Execute(dataCtx, kb); err != nil {
		return nil, err
	}

	// 5. Convert results back to domain entities
	return e.convertToTransactions(journal), nil
}

// prepareDataContext converts domain entity to Grule-compatible format
func (e *GruleEngine) prepareDataContext(walletTx *transformation.WalletTransaction) (ast.IDataContext, *GruleJournal) {
	// Convert domain entity to Grule input format
	// ...
	return nil, nil
}
