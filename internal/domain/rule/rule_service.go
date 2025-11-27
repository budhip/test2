package rule

import (
	"context"
	"fmt"
)

type RuleValidator interface {
	ValidateGruleSyntax(content string, ruleName string) error
}

type RuleService struct {
	validator RuleValidator
}

func NewRuleService(validator RuleValidator) *RuleService {
	return &RuleService{
		validator: validator,
	}
}

func (s *RuleService) ValidateRuleContent(ctx context.Context, content Content, ruleName RuleName) error {
	return s.validator.ValidateGruleSyntax(content.String(), ruleName.String())
}

func (s *RuleService) CheckDuplicateRule(ctx context.Context, repo RuleRepository, name RuleName, env Environment, version Version) error {
	existing, err := repo.FindByNameEnvVersion(ctx, name, env, version)
	if err != nil {
		return nil
	}

	if existing != nil {
		return fmt.Errorf("rule already exists: %s/%s/%s", env.String(), name.String(), version.String())
	}

	return nil
}
