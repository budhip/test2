package transformation

import (
	"context"
	"fmt"

	"bitbucket.org/Amartha/go-megatron/internal/domain/rule"
)

type TransformationService struct {
	engine     TransformationEngine
	ruleRepo   rule.RuleRepository
	validation *TransformationValidator
}

func NewTransformationService(
	engine TransformationEngine,
	ruleRepo rule.RuleRepository,
	validation *TransformationValidator,
) *TransformationService {
	return &TransformationService{
		engine:     engine,
		ruleRepo:   ruleRepo,
		validation: validation,
	}
}

func (s *TransformationService) TransformWalletTransaction(
	ctx context.Context,
	wt *WalletTransaction,
) ([]*Transaction, error) {

	if err := s.validation.ValidateWalletTransaction(wt); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	transformableAmounts := wt.GetAmountsForTransformation()

	if len(transformableAmounts) == 0 {
		return nil, ErrNoAmountToTransform
	}

	var allTransactions []*Transaction

	for _, transformable := range transformableAmounts {
		ruleEntity, err := s.ruleRepo.FindByTransactionType(ctx, transformable.TransactionType.String())
		if err != nil {
			return nil, fmt.Errorf("rule not found for %s: %w",
				transformable.TransactionType.String(), err)
		}

		transactions, err := s.engine.Transform(ctx, wt, ruleEntity.Content().String())
		if err != nil {
			return nil, fmt.Errorf("transformation failed for %s: %w",
				transformable.TransactionType.String(), err)
		}

		for _, tx := range transactions {
			if err := s.validation.ValidateTransaction(tx); err != nil {
				return nil, fmt.Errorf("invalid transaction generated: %w", err)
			}
		}

		allTransactions = append(allTransactions, transactions...)
	}

	wt.MarkAsTransformed()

	return allTransactions, nil
}
