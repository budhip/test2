package rule

import (
	"context"
	"fmt"

	"bitbucket.org/Amartha/go-megatron/internal/domain/rule"
)

type UpdateRuleCommand struct {
	RuleID  string
	Content string
}

type UpdateRuleHandler struct {
	ruleRepo    rule.RuleRepository
	ruleService *rule.RuleService
}

func NewUpdateRuleHandler(
	ruleRepo rule.RuleRepository,
	ruleService *rule.RuleService,
) *UpdateRuleHandler {
	return &UpdateRuleHandler{
		ruleRepo:    ruleRepo,
		ruleService: ruleService,
	}
}

func (h *UpdateRuleHandler) Handle(ctx context.Context, cmd UpdateRuleCommand) error {
	ruleID, err := rule.ParseRuleID(cmd.RuleID)
	if err != nil {
		return fmt.Errorf("invalid rule ID: %w", err)
	}

	existingRule, err := h.ruleRepo.FindByID(ctx, ruleID)
	if err != nil {
		return fmt.Errorf("rule not found: %w", err)
	}

	if err := existingRule.UpdateContent(cmd.Content); err != nil {
		return fmt.Errorf("failed to update content: %w", err)
	}

	if err := h.ruleService.ValidateRuleContent(ctx, existingRule.Content(), existingRule.Name()); err != nil {
		return fmt.Errorf("invalid rule content: %w", err)
	}

	if err := h.ruleRepo.Update(ctx, existingRule); err != nil {
		return fmt.Errorf("failed to update rule: %w", err)
	}

	return nil
}
