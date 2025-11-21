package service

import (
	"bitbucket.org/Amartha/go-megatron/internal/acuanrepository"
	"bitbucket.org/Amartha/go-megatron/internal/megatron"
	"bitbucket.org/Amartha/go-megatron/internal/transformer"
	"context"
)

type TransformService interface {
	Transform(ctx context.Context, req megatron.TransformRequest) (*megatron.TransformResponse, error)
	BatchTransform(ctx context.Context, req megatron.BatchTransformRequest) (*megatron.BatchTransformResponse, error)
}

type transformService struct {
	engine *transformer.Engine
}

func NewTransformService(ruleRepo acuanrepository.RuleRepository) TransformService {
	engine := transformer.NewEngine(ruleRepo, transformer.Config{
		DefaultCurrency: "IDR",
	})

	return &transformService{
		engine: engine,
	}
}

func (s *transformService) Transform(ctx context.Context, req megatron.TransformRequest) (*megatron.TransformResponse, error) {
	return s.engine.Transform(ctx, req)
}

func (s *transformService) BatchTransform(ctx context.Context, req megatron.BatchTransformRequest) (*megatron.BatchTransformResponse, error) {
	var allTransactions []megatron.TransactionOutput
	var errors []megatron.TransformError
	var lastMetadata megatron.TransformMetadata

	for _, transform := range req.Transforms {
		singleReq := megatron.TransformRequest{
			ParentTransaction: req.ParentTransaction,
			Amount:            transform.Amount,
			TransactionType:   transform.TransactionType,
			Context:           req.Context,
		}

		resp, err := s.engine.Transform(ctx, singleReq)
		if err != nil {
			errors = append(errors, megatron.TransformError{
				TransactionType: transform.TransactionType,
				Error:           err.Error(),
				Code:            "TRANSFORM_ERROR",
			})
			continue
		}

		allTransactions = append(allTransactions, resp.Transactions...)
		lastMetadata = resp.Metadata
	}

	return &megatron.BatchTransformResponse{
		Transactions: allTransactions,
		Errors:       errors,
		Metadata:     lastMetadata,
	}, nil
}
