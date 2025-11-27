package rule

import (
	"context"
	"fmt"

	"bitbucket.org/Amartha/go-megatron/internal/application/dto"
	"bitbucket.org/Amartha/go-megatron/internal/domain/rule"
)

type GetRuleQuery struct {
	Name    string
	Env     string
	Version string
}

type GetRuleHandler struct {
	ruleRepo rule.RuleRepository
}

func NewGetRuleHandler(ruleRepo rule.RuleRepository) *GetRuleHandler {
	return &GetRuleHandler{
		ruleRepo: ruleRepo,
	}
}

func (h *GetRuleHandler) Handle(ctx context.Context, query GetRuleQuery) (*dto.RuleDTO, error) {
	ruleName, err := rule.NewRuleName(query.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid rule name: %w", err)
	}

	env, err := rule.NewEnvironment(query.Env)
	if err != nil {
		return nil, fmt.Errorf("invalid environment: %w", err)
	}

	version, err := rule.NewVersion(query.Version)
	if err != nil {
		return nil, fmt.Errorf("invalid version: %w", err)
	}

	ruleEntity, err := h.ruleRepo.FindByNameEnvVersion(ctx, ruleName, env, version)
	if err != nil {
		return nil, fmt.Errorf("rule not found: %w", err)
	}

	return dto.FromRuleEntity(ruleEntity), nil
}
