package services

import (
	"bitbucket.org/Amartha/go-megatron/internal/common/grule"
	"bitbucket.org/Amartha/go-megatron/internal/models"
	"bitbucket.org/Amartha/go-megatron/internal/repositories"
	"context"
)

type TransformService interface {
	TransformWalletTransaction(ctx context.Context, req models.WalletTransactionRequest) (*models.WalletTransactionResponse, error)
}

type transformService struct {
	gruleEngine *grule.Engine
}

func NewTransformService(ruleRepo repositories.AcuanRuleRepository) TransformService {
	return &transformService{
		gruleEngine: grule.NewEngine(ruleRepo),
	}
}

func (s *transformService) TransformWalletTransaction(ctx context.Context, req models.WalletTransactionRequest) (*models.WalletTransactionResponse, error) {
	return s.gruleEngine.TransformWalletTransaction(ctx, req)
}
