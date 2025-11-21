package repositories

import (
	"bitbucket.org/Amartha/go-megatron/internal/models"
	"context"
	"database/sql"
	"fmt"
)

type AcuanRuleRepository interface {
	GetActiveRule(ctx context.Context, transactionType string) (*models.AcuanRule, error)
}

type acuanRuleRepository struct {
	db *sql.DB
}

func NewAcuanRuleRepository(db *sql.DB) AcuanRuleRepository {
	return &acuanRuleRepository{db: db}
}

func (r *acuanRuleRepository) GetActiveRule(ctx context.Context, transactionType string) (*models.AcuanRule, error) {
	query := `
		SELECT r.id, r.transaction_type, r.rule_name, r.description, r.version, r.is_active,
		       rv.config, r.tags, r.created_at, r.updated_at
		FROM transformation_rules r
		JOIN rule_versions rv ON r.id = rv.rule_id AND r.version = rv.version
		WHERE r.transaction_type = $1 AND r.is_active = true
	`

	rule := &models.AcuanRule{}
	err := r.db.QueryRowContext(ctx, query, transactionType).Scan(
		&rule.ID,
		&rule.TransactionType,
		&rule.RuleName,
		&rule.Description,
		&rule.Version,
		&rule.IsActive,
		&rule.Config,
		&rule.Tags,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no active rule found for transaction type: %s", transactionType)
	}

	return rule, err
}
