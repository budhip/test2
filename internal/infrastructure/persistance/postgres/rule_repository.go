package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"go-megatron/internal/domain/rule"
	"go-megatron/internal/infrastructure/persistence/postgres/mappers"
)

type RuleRepository struct {
	db *sql.DB
}

func NewRuleRepository(db *sql.DB) *RuleRepository {
	return &RuleRepository{db: db}
}

func (r *RuleRepository) FindByID(ctx context.Context, id rule.RuleID) (*rule.Rule, error) {
	query := `
		SELECT id, name, env, version, content, is_active, created_at, updated_at, created_by, updated_by
		FROM rules
		WHERE id = $1 AND is_active = true
	`

	var dbRule DBRule
	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&dbRule.ID,
		&dbRule.Name,
		&dbRule.Env,
		&dbRule.Version,
		&dbRule.Content,
		&dbRule.IsActive,
		&dbRule.CreatedAt,
		&dbRule.UpdatedAt,
		&dbRule.CreatedBy,
		&dbRule.UpdatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("rule not found")
	}
	if err != nil {
		return nil, err
	}

	return mappers.ToDomainRule(dbRule)
}

func (r *RuleRepository) FindByNameEnvVersion(ctx context.Context, name rule.RuleName, env rule.Environment, version rule.Version) (*rule.Rule, error) {
	query := `
		SELECT id, name, env, version, content, is_active, created_at, updated_at, created_by, updated_by
		FROM rules
		WHERE name = $1 AND env = $2 AND version = $3 AND is_active = true
	`

	var dbRule DBRule
	err := r.db.QueryRowContext(ctx, query, name.String(), env.String(), version.String()).Scan(
		&dbRule.ID,
		&dbRule.Name,
		&dbRule.Env,
		&dbRule.Version,
		&dbRule.Content,
		&dbRule.IsActive,
		&dbRule.CreatedAt,
		&dbRule.UpdatedAt,
		&dbRule.CreatedBy,
		&dbRule.UpdatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("rule not found: %s/%s/%s", env.String(), name.String(), version.String())
	}
	if err != nil {
		return nil, err
	}

	return mappers.ToDomainRule(dbRule)
}

func (r *RuleRepository) FindLatest(ctx context.Context, name rule.RuleName, env rule.Environment) (*rule.Rule, error) {
	query := `
		SELECT id, name, env, version, content, is_active, created_at, updated_at, created_by, updated_by
		FROM rules
		WHERE name = $1 AND env = $2 AND is_active = true
		ORDER BY created_at DESC
		LIMIT 1
	`

	var dbRule DBRule
	err := r.db.QueryRowContext(ctx, query, name.String(), env.String()).Scan(
		&dbRule.ID,
		&dbRule.Name,
		&dbRule.Env,
		&dbRule.Version,
		&dbRule.Content,
		&dbRule.IsActive,
		&dbRule.CreatedAt,
		&dbRule.UpdatedAt,
		&dbRule.CreatedBy,
		&dbRule.UpdatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("rule not found: %s/%s", env.String(), name.String())
	}
	if err != nil {
		return nil, err
	}

	return mappers.ToDomainRule(dbRule)
}

func (r *RuleRepository) FindAllByEnv(ctx context.Context, env rule.Environment) ([]*rule.Rule, error) {
	query := `
		SELECT id, name, env, version, content, is_active, created_at, updated_at, created_by, updated_by
		FROM rules
		WHERE env = $1 AND is_active = true
		ORDER BY name, version
	`

	rows, err := r.db.QueryContext(ctx, query, env.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*rule.Rule

	for rows.Next() {
		var dbRule DBRule
		err := rows.Scan(
			&dbRule.ID,
			&dbRule.Name,
			&dbRule.Env,
			&dbRule.Version,
			&dbRule.Content,
			&dbRule.IsActive,
			&dbRule.CreatedAt,
			&dbRule.UpdatedAt,
			&dbRule.CreatedBy,
			&dbRule.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}

		ruleEntity, err := mappers.ToDomainRule(dbRule)
		if err != nil {
			return nil, err
		}

		rules = append(rules, ruleEntity)
	}

	return rules, rows.Err()
}

func (r *RuleRepository) FindByTransactionType(ctx context.Context, transactionType string) (*rule.Rule, error) {
	query := `
		SELECT r.id, r.name, r.env, r.version, r.content, r.is_active, r.created_at, r.updated_at, r.created_by, r.updated_by
		FROM rules r
		INNER JOIN transformation_rules tr ON tr.rule_id = r.id
		WHERE tr.transaction_type = $1 AND tr.is_active = true AND r.is_active = true
		LIMIT 1
	`

	var dbRule DBRule
	err := r.db.QueryRowContext(ctx, query, transactionType).Scan(
		&dbRule.ID,
		&dbRule.Name,
		&dbRule.Env,
		&dbRule.Version,
		&dbRule.Content,
		&dbRule.IsActive,
		&dbRule.CreatedAt,
		&dbRule.UpdatedAt,
		&dbRule.CreatedBy,
		&dbRule.UpdatedBy,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no rule found for transaction type: %s", transactionType)
	}
	if err != nil {
		return nil, err
	}

	return mappers.ToDomainRule(dbRule)
}

func (r *RuleRepository) Save(ctx context.Context, ruleEntity *rule.Rule) error {
	dbRule := mappers.ToDBRule(ruleEntity)

	query := `
		INSERT INTO rules (id, name, env, version, content, is_active, created_at, updated_at, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		dbRule.ID,
		dbRule.Name,
		dbRule.Env,
		dbRule.Version,
		dbRule.Content,
		dbRule.IsActive,
		dbRule.CreatedAt,
		dbRule.UpdatedAt,
		dbRule.CreatedBy,
		dbRule.UpdatedBy,
	)

	return err
}

func (r *RuleRepository) Update(ctx context.Context, ruleEntity *rule.Rule) error {
	dbRule := mappers.ToDBRule(ruleEntity)

	query := `
		UPDATE rules
		SET content = $1, version = $2, updated_at = $3, updated_by = $4
		WHERE id = $5 AND is_active = true
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		dbRule.Content,
		dbRule.Version,
		dbRule.UpdatedAt,
		dbRule.UpdatedBy,
		dbRule.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("rule not found or already deleted: %s", dbRule.ID)
	}

	return nil
}

func (r *RuleRepository) Delete(ctx context.Context, id rule.RuleID) error {
	query := `
		UPDATE rules
		SET is_active = false, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("rule not found: %s", id.String())
	}

	return nil
}
