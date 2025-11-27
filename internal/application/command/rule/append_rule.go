package rule

import (
	"context"
	"fmt"

	"bitbucket.org/Amartha/go-megatron/internal/domain/rule"
)

type AppendRuleCommand struct {
	RuleID      string
	Content     string
	InsertMode  string
	AutoVersion bool
	VersionBump string
}

type AppendRuleHandler struct {
	ruleRepo    rule.RuleRepository
	ruleService *rule.RuleService
}

func NewAppendRuleHandler(
	ruleRepo rule.RuleRepository,
	ruleService *rule.RuleService,
) *AppendRuleHandler {
	return &AppendRuleHandler{
		ruleRepo:    ruleRepo,
		ruleService: ruleService,
	}
}

func (h *AppendRuleHandler) Handle(ctx context.Context, cmd AppendRuleCommand) error {
	ruleID, err := rule.ParseRuleID(cmd.RuleID)
	if err != nil {
		return fmt.Errorf("invalid rule ID: %w", err)
	}

	existingRule, err := h.ruleRepo.FindByID(ctx, ruleID)
	if err != nil {
		return fmt.Errorf("rule not found: %w", err)
	}

	insertMode := rule.InsertMode(cmd.InsertMode)

	if err := existingRule.AppendContent(cmd.Content, insertMode); err != nil {
		return fmt.Errorf("failed to append content: %w", err)
	}

	if cmd.AutoVersion {
		bumpType := rule.VersionBumpType(cmd.VersionBump)
		if err := existingRule.BumpVersion(bumpType); err != nil {
			return fmt.Errorf("failed to bump version: %w", err)
		}
	}

	if err := h.ruleService.ValidateRuleContent(ctx, existingRule.Content(), existingRule.Name()); err != nil {
		return fmt.Errorf("invalid merged content: %w", err)
	}

	if err := h.ruleRepo.Update(ctx, existingRule); err != nil {
		return fmt.Errorf("failed to update rule: %w", err)
	}

	return nil
}
