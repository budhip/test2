package repository

import (
	"context"
	"database/sql"
	"fmt"

	xlog "bitbucket.org/Amartha/go-x/log"
)

// RuleRepository defines the interface for rule data access
type RuleRepository interface {
	GetRule(ctx context.Context, name, env, version string) (*Rule, error)
	GetRuleByID(ctx context.Context, id int64) (*Rule, error)
	CreateRule(ctx context.Context, rule *Rule) error
	UpdateRule(ctx context.Context, rule *Rule) error
	DeleteRule(ctx context.Context, id int64) error
	ListRules(ctx context.Context, env string) ([]*Rule, error)
	GetLatestRule(ctx context.Context, name, env string) (*Rule, error)
}

type ruleRepository struct {
	db *sql.DB
}

// NewRuleRepository creates a new rule repository
func NewRuleRepository(db *sql.DB) RuleRepository {
	return &ruleRepository{db: db}
}

// GetRule retrieves a rule by name, environment, and version
func (r *ruleRepository) GetRule(ctx context.Context, name, env, version string) (*Rule, error) {
	query := `
		SELECT id, name, env, version, content, is_active, created_at, updated_at
		FROM rules
		WHERE name = $1 AND env = $2 AND version = $3 AND is_active = true
	`

	xlog.Info(ctx, "[RULE_REPOSITORY] Getting rule",
		xlog.String("name", name),
		xlog.String("env", env),
		xlog.String("version", version))

	var rule Rule
	err := r.db.QueryRowContext(ctx, query, name, env, version).Scan(
		&rule.ID,
		&rule.Name,
		&rule.Env,
		&rule.Version,
		&rule.Content,
		&rule.IsActive,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("rule not found: %s/%s/%s", env, name, version)
	}

	if err != nil {
		xlog.Error(ctx, "[RULE_REPOSITORY] Failed to get rule", xlog.Err(err))
		return nil, fmt.Errorf("error getting rule: %w", err)
	}

	xlog.Info(ctx, "[RULE_REPOSITORY] Rule retrieved successfully", xlog.Int64("id", rule.ID))
	return &rule, nil
}

// GetRuleByID retrieves a rule by its ID
func (r *ruleRepository) GetRuleByID(ctx context.Context, id int64) (*Rule, error) {
	query := `
		SELECT id, name, env, version, content, is_active, created_at, updated_at
		FROM rules
		WHERE id = $1 AND is_active = true
	`

	xlog.Info(ctx, "[RULE_REPOSITORY] Getting rule by ID", xlog.Int64("id", id))

	var rule Rule
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rule.ID,
		&rule.Name,
		&rule.Env,
		&rule.Version,
		&rule.Content,
		&rule.IsActive,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("rule not found with id: %d", id)
	}

	if err != nil {
		xlog.Error(ctx, "[RULE_REPOSITORY] Failed to get rule by ID", xlog.Err(err))
		return nil, fmt.Errorf("error getting rule by ID: %w", err)
	}

	xlog.Info(ctx, "[RULE_REPOSITORY] Rule retrieved successfully",
		xlog.Int64("id", rule.ID),
		xlog.String("name", rule.Name))
	return &rule, nil
}

// CreateRule creates a new rule in the database
func (r *ruleRepository) CreateRule(ctx context.Context, rule *Rule) error {
	query := `
		INSERT INTO rules (name, env, version, content, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	xlog.Info(ctx, "[RULE_REPOSITORY] Creating rule",
		xlog.String("name", rule.Name),
		xlog.String("env", rule.Env),
		xlog.String("version", rule.Version))

	err := r.db.QueryRowContext(
		ctx,
		query,
		rule.Name,
		rule.Env,
		rule.Version,
		rule.Content,
		rule.IsActive,
	).Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)

	if err != nil {
		xlog.Error(ctx, "[RULE_REPOSITORY] Failed to create rule", xlog.Err(err))
		return fmt.Errorf("error creating rule: %w", err)
	}

	xlog.Info(ctx, "[RULE_REPOSITORY] Rule created successfully", xlog.Int64("id", rule.ID))
	return nil
}

// UpdateRule updates an existing rule
func (r *ruleRepository) UpdateRule(ctx context.Context, rule *Rule) error {
	query := `
		UPDATE rules
		SET content = $1, version = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3 AND is_active = true
	`

	xlog.Info(ctx, "[RULE_REPOSITORY] Updating rule",
		xlog.Int64("id", rule.ID),
		xlog.String("version", rule.Version))

	result, err := r.db.ExecContext(ctx, query, rule.Content, rule.Version, rule.ID)
	if err != nil {
		xlog.Error(ctx, "[RULE_REPOSITORY] Failed to update rule", xlog.Err(err))
		return fmt.Errorf("error updating rule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("rule not found or already deleted: %d", rule.ID)
	}

	xlog.Info(ctx, "[RULE_REPOSITORY] Rule updated successfully",
		xlog.Int64("id", rule.ID),
		xlog.String("new_version", rule.Version))
	return nil
}

// DeleteRule soft deletes a rule by setting is_active to false
func (r *ruleRepository) DeleteRule(ctx context.Context, id int64) error {
	query := `
		UPDATE rules
		SET is_active = false, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	xlog.Info(ctx, "[RULE_REPOSITORY] Deleting rule", xlog.Int64("id", id))

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		xlog.Error(ctx, "[RULE_REPOSITORY] Failed to delete rule", xlog.Err(err))
		return fmt.Errorf("error deleting rule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("rule not found: %d", id)
	}

	xlog.Info(ctx, "[RULE_REPOSITORY] Rule deleted successfully", xlog.Int64("id", id))
	return nil
}

// ListRules retrieves all active rules for a specific environment
func (r *ruleRepository) ListRules(ctx context.Context, env string) ([]*Rule, error) {
	query := `
		SELECT id, name, env, version, content, is_active, created_at, updated_at
		FROM rules
		WHERE env = $1 AND is_active = true
		ORDER BY name, version
	`

	xlog.Info(ctx, "[RULE_REPOSITORY] Listing rules", xlog.String("env", env))

	rows, err := r.db.QueryContext(ctx, query, env)
	if err != nil {
		xlog.Error(ctx, "[RULE_REPOSITORY] Failed to list rules", xlog.Err(err))
		return nil, fmt.Errorf("error listing rules: %w", err)
	}
	defer rows.Close()

	var rules []*Rule
	for rows.Next() {
		var rule Rule
		err := rows.Scan(
			&rule.ID,
			&rule.Name,
			&rule.Env,
			&rule.Version,
			&rule.Content,
			&rule.IsActive,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		)
		if err != nil {
			xlog.Error(ctx, "[RULE_REPOSITORY] Failed to scan rule", xlog.Err(err))
			return nil, fmt.Errorf("error scanning rule: %w", err)
		}
		rules = append(rules, &rule)
	}

	if err = rows.Err(); err != nil {
		xlog.Error(ctx, "[RULE_REPOSITORY] Error iterating rules", xlog.Err(err))
		return nil, fmt.Errorf("error iterating rules: %w", err)
	}

	xlog.Info(ctx, "[RULE_REPOSITORY] Rules listed successfully", xlog.Int("count", len(rules)))
	return rules, nil
}

func (r *ruleRepository) GetLatestRule(ctx context.Context, name, env string) (*Rule, error) {
	query := `
		SELECT id, name, env, version, content, is_active, created_at, updated_at
		FROM rules
		WHERE name = $1 AND env = $2 AND is_active = true
		ORDER BY version DESC
		LIMIT 1
	`

	xlog.Info(ctx, "[RULE_REPOSITORY] Getting latest rule",
		xlog.String("name", name),
		xlog.String("env", env))

	var rule Rule
	err := r.db.QueryRowContext(ctx, query, name, env).Scan(
		&rule.ID,
		&rule.Name,
		&rule.Env,
		&rule.Version,
		&rule.Content,
		&rule.IsActive,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("rule not found: %s/%s", env, name)
	}

	if err != nil {
		xlog.Error(ctx, "[RULE_REPOSITORY] Failed to get latest rule", xlog.Err(err))
		return nil, fmt.Errorf("error getting latest rule: %w", err)
	}

	xlog.Info(ctx, "[RULE_REPOSITORY] Latest rule retrieved successfully",
		xlog.Int64("id", rule.ID),
		xlog.String("version", rule.Version))
	return &rule, nil
}
