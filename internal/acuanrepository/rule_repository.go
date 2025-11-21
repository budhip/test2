package acuanrepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type RuleRepository interface {
	GetActiveRule(ctx context.Context, transactionType string) (*Rule, error)
	GetRuleByID(ctx context.Context, id string) (*Rule, error)
	GetRuleByTransactionType(ctx context.Context, transactionType string) (*Rule, error)
	CreateRule(ctx context.Context, req CreateRuleRequest) (*Rule, error)
	UpdateRule(ctx context.Context, transactionType string, req UpdateRuleRequest) (*Rule, error)
	DeactivateRule(ctx context.Context, transactionType string) error
	ListRules(ctx context.Context, req ListRulesRequest) ([]Rule, int, error)
	GetRuleVersions(ctx context.Context, ruleID string) ([]RuleVersion, error)
}

type ruleRepository struct {
	db *sql.DB
}

func NewRuleRepository(db *sql.DB) RuleRepository {
	return &ruleRepository{db: db}
}

type Rule struct {
	ID              string
	TransactionType string
	RuleName        string
	Description     string
	Version         int
	IsActive        bool
	Config          json.RawMessage
	Tags            json.RawMessage
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CreatedBy       string
}

type RuleVersion struct {
	ID          string
	RuleID      string
	Version     int
	Config      json.RawMessage
	CreatedAt   time.Time
	CreatedBy   string
	ChangeNotes string
}

type CreateRuleRequest struct {
	TransactionType string
	RuleName        string
	Description     string
	Config          json.RawMessage
	Tags            json.RawMessage
	CreatedBy       string
}

type UpdateRuleRequest struct {
	Config      json.RawMessage
	ChangeNotes string
	UpdatedBy   string
}

type ListRulesRequest struct {
	TransactionTypes []string
	IsActive         *bool
	Tags             map[string]interface{}
	Limit            int
	Offset           int
}

func (r *ruleRepository) GetActiveRule(ctx context.Context, transactionType string) (*Rule, error) {
	query := `
		SELECT r.id, r.transaction_type, r.rule_name, r.description, r.version, r.is_active,
		       rv.config, r.tags, r.created_at, r.updated_at
		FROM transformation_rules r
		JOIN rule_versions rv ON r.id = rv.rule_id AND r.version = rv.version
		WHERE r.transaction_type = $1 AND r.is_active = true
	`

	rule := &Rule{}
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

func (r *ruleRepository) GetRuleByTransactionType(ctx context.Context, transactionType string) (*Rule, error) {
	query := `
		SELECT r.id, r.transaction_type, r.rule_name, r.description, r.version, r.is_active,
		       rv.config, r.tags, r.created_at, r.updated_at
		FROM transformation_rules r
		JOIN rule_versions rv ON r.id = rv.rule_id AND r.version = rv.version
		WHERE r.transaction_type = $1
		ORDER BY r.version DESC
		LIMIT 1
	`

	rule := &Rule{}
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
		return nil, fmt.Errorf("rule not found for transaction type: %s", transactionType)
	}

	return rule, err
}

func (r *ruleRepository) CreateRule(ctx context.Context, req CreateRuleRequest) (*Rule, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Create rule
	ruleID := uuid.New().String()
	now := time.Now()

	insertRule := `
		INSERT INTO transformation_rules 
		(id, transaction_type, rule_name, description, version, is_active, created_at, updated_at, tags)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = tx.ExecContext(ctx, insertRule,
		ruleID, req.TransactionType, req.RuleName, req.Description, 1, true, now, now, req.Tags)
	if err != nil {
		return nil, err
	}

	// Create version
	versionID := uuid.New().String()
	insertVersion := `
		INSERT INTO rule_versions
		(id, rule_id, version, config, created_at, change_notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = tx.ExecContext(ctx, insertVersion,
		versionID, ruleID, 1, req.Config, now, "Initial version")
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetRuleByID(ctx, ruleID)
}

func (r *ruleRepository) UpdateRule(ctx context.Context, transactionType string, req UpdateRuleRequest) (*Rule, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Get current rule
	rule, err := r.GetRuleByTransactionType(ctx, transactionType)
	if err != nil {
		return nil, err
	}

	newVersion := rule.Version + 1
	now := time.Now()

	// Update rule version
	updateRule := `
		UPDATE transformation_rules
		SET version = $1, updated_at = $2
		WHERE id = $3
	`

	_, err = tx.ExecContext(ctx, updateRule, newVersion, now, rule.ID)
	if err != nil {
		return nil, err
	}

	// Insert new version
	versionID := uuid.New().String()
	insertVersion := `
		INSERT INTO rule_versions
		(id, rule_id, version, config, created_at, change_notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = tx.ExecContext(ctx, insertVersion,
		versionID, rule.ID, newVersion, req.Config, now, req.UpdatedBy, req.ChangeNotes)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetRuleByID(ctx, rule.ID)
}

func (r *ruleRepository) GetRuleByID(ctx context.Context, id string) (*Rule, error) {
	query := `
		SELECT r.id, r.transaction_type, r.rule_name, r.description, r.version, r.is_active,
		       rv.config, r.tags, r.created_at, r.updated_at
		FROM transformation_rules r
		JOIN rule_versions rv ON r.id = rv.rule_id AND r.version = rv.version
		WHERE r.id = $1
	`

	rule := &Rule{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
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

	return rule, err
}

func (r *ruleRepository) DeactivateRule(ctx context.Context, transactionType string) error {
	query := `
		UPDATE transformation_rules
		SET is_active = false, updated_at = $1
		WHERE transaction_type = $2
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), transactionType)
	return err
}

func (r *ruleRepository) ListRules(ctx context.Context, req ListRulesRequest) ([]Rule, int, error) {
	// Build query
	query := `
		SELECT r.id, r.transaction_type, r.rule_name, r.description, r.version, r.is_active,
		       rv.config, r.tags, r.created_at, r.updated_at
		FROM transformation_rules r
		JOIN rule_versions rv ON r.id = rv.rule_id AND r.version = rv.version
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 1

	if len(req.TransactionTypes) > 0 {
		query += fmt.Sprintf(" AND r.transaction_type = ANY($%d)", argCount)
		args = append(args, pq.Array(req.TransactionTypes))
		argCount++
	}

	if req.IsActive != nil {
		query += fmt.Sprintf(" AND r.is_active = $%d", argCount)
		args = append(args, *req.IsActive)
		argCount++
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) as count_query", query)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Add pagination
	query += fmt.Sprintf(" ORDER BY r.created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, req.Limit, req.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var rules []Rule
	for rows.Next() {
		var rule Rule
		err := rows.Scan(
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
		if err != nil {
			return nil, 0, err
		}
		rules = append(rules, rule)
	}

	return rules, total, nil
}

func (r *ruleRepository) GetRuleVersions(ctx context.Context, ruleID string) ([]RuleVersion, error) {
	query := `
		SELECT id, rule_id, version, config, created_at, change_notes
		FROM rule_versions
		WHERE rule_id = $1
		ORDER BY version DESC
	`

	rows, err := r.db.QueryContext(ctx, query, ruleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []RuleVersion
	for rows.Next() {
		var version RuleVersion
		err := rows.Scan(
			&version.ID,
			&version.RuleID,
			&version.Version,
			&version.Config,
			&version.CreatedAt,
			&version.ChangeNotes,
		)
		if err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}

	return versions, nil
}
