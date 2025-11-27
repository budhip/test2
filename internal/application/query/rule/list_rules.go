package rule

import (
	"context"
	"fmt"

	"bitbucket.org/Amartha/go-megatron/internal/application/dto"
	"bitbucket.org/Amartha/go-megatron/internal/domain/rule"
)

type ListRulesQuery struct {
	Env string
}

type ListRulesHandler struct {
	ruleRepo rule.RuleRepository
}

func NewListRulesHandler(ruleRepo rule.RuleRepository) *ListRulesHandler {
	return &ListRulesHandler{
		ruleRepo: ruleRepo,
	}
}

func (h *ListRulesHandler) Handle(ctx context.Context, query ListRulesQuery) ([]*dto.RuleDTO, error) {
	env, err := rule.NewEnvironment(query.Env)
	if err != nil {
		return nil, fmt.Errorf("invalid environment: %w", err)
	}

	rules, err := h.ruleRepo.FindAllByEnv(ctx, env)
	if err != nil {
		return nil, fmt.Errorf("failed to list rules: %w", err)
	}

	result := make([]*dto.RuleDTO, len(rules))
	for i, r := range rules {
		result[i] = dto.FromRuleEntity(r)
	}

	return result, nil
}
