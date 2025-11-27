package rule

import (
	"context"
	"fmt"

	"bitbucket.org/Amartha/go-megatron/internal/domain/rule"
)

type DeleteRuleCommand struct {
	RuleID string
}

type DeleteRuleHandler struct {
	ruleRepo rule.RuleRepository
}

func NewDeleteRuleHandler(ruleRepo rule.RuleRepository) *DeleteRuleHandler {
	return &DeleteRuleHandler{
		ruleRepo: ruleRepo,
	}
}

func (h *DeleteRuleHandler) Handle(ctx context.Context, cmd DeleteRuleCommand) error {
	ruleID, err := rule.ParseRuleID(cmd.RuleID)
	if err != nil {
		return fmt.Errorf("invalid rule ID: %w", err)
	}

	if err := h.ruleRepo.Delete(ctx, ruleID); err != nil {
		return fmt.Errorf("failed to delete rule: %w", err)
	}

	return nil
}
