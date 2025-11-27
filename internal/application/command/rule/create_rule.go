package rule

import (
	"context"
	"fmt"

	"bitbucket.org/Amartha/go-megatron/internal/application/dto"
	"bitbucket.org/Amartha/go-megatron/internal/domain/rule"
)

type CreateRuleCommand struct {
	Name    string
	Env     string
	Version string
	Content string
}

type CreateRuleHandler struct {
	ruleRepo    rule.RuleRepository
	ruleService *rule.RuleService
}

func NewCreateRuleHandler(
	ruleRepo rule.RuleRepository,
	ruleService *rule.RuleService,
) *CreateRuleHandler {
	return &CreateRuleHandler{
		ruleRepo:    ruleRepo,
		ruleService: ruleService,
	}
}

func (h *CreateRuleHandler) Handle(ctx context.Context, cmd CreateRuleCommand) (*dto.RuleDTO, error) {
	newRule, err := rule.NewRule(cmd.Name, cmd.Env, cmd.Version, cmd.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to create rule: %w", err)
	}

	if err := h.ruleService.ValidateRuleContent(ctx, newRule.Content(), newRule.Name()); err != nil {
		return nil, fmt.Errorf("invalid rule content: %w", err)
	}

	if err := h.ruleService.CheckDuplicateRule(ctx, h.ruleRepo, newRule.Name(), newRule.Env(), newRule.Version()); err != nil {
		return nil, err
	}

	if err := h.ruleRepo.Save(ctx, newRule); err != nil {
		return nil, fmt.Errorf("failed to save rule: %w", err)
	}

	return dto.FromRuleEntity(newRule), nil
}
