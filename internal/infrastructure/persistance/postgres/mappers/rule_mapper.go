package mappers

import (
	"go-megatron/internal/domain/rule"
	"go-megatron/internal/infrastructure/persistence/postgres"
)

func ToDomainRule(dbRule postgres.DBRule) (*rule.Rule, error) {
	ruleID, err := rule.ParseRuleID(dbRule.ID)
	if err != nil {
		return nil, err
	}

	ruleName, err := rule.NewRuleName(dbRule.Name)
	if err != nil {
		return nil, err
	}

	env, err := rule.NewEnvironment(dbRule.Env)
	if err != nil {
		return nil, err
	}

	version, err := rule.NewVersion(dbRule.Version)
	if err != nil {
		return nil, err
	}

	content, err := rule.NewContent(dbRule.Content)
	if err != nil {
		return nil, err
	}

	return rule.ReconstructRule(
		ruleID,
		ruleName,
		env,
		version,
		content,
		dbRule.IsActive,
		dbRule.CreatedAt,
		dbRule.UpdatedAt,
		dbRule.CreatedBy,
		dbRule.UpdatedBy,
	), nil
}

func ToDBRule(r *rule.Rule) postgres.DBRule {
	return postgres.DBRule{
		ID:        r.ID().String(),
		Name:      r.Name().String(),
		Env:       r.Env().String(),
		Version:   r.Version().String(),
		Content:   r.Content().String(),
		IsActive:  r.IsActive(),
		CreatedAt: r.CreatedAt(),
		UpdatedAt: r.UpdatedAt(),
		CreatedBy: r.CreatedBy(),
		UpdatedBy: r.UpdatedBy(),
	}
}
