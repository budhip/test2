package rule

import "context"

type RuleRepository interface {
	FindByID(ctx context.Context, id RuleID) (*Rule, error)
	FindByNameEnvVersion(ctx context.Context, name RuleName, env Environment, version Version) (*Rule, error)
	FindLatest(ctx context.Context, name RuleName, env Environment) (*Rule, error)
	FindAllByEnv(ctx context.Context, env Environment) ([]*Rule, error)
	FindByTransactionType(ctx context.Context, transactionType string) (*Rule, error)

	Save(ctx context.Context, rule *Rule) error
	Update(ctx context.Context, rule *Rule) error
	Delete(ctx context.Context, id RuleID) error
}
