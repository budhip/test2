package rules

import (
	"context"
	"fmt"
	"time"

	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/hyperjumptech/grule-rule-engine/pkg"
)

// RuleLoader is interface to get rule resource
type RuleLoader interface {
	LoadRule(name, env, version string) (pkg.Resource, error)
}

// FileRuleLoader loads rules from files (legacy/fallback)
type FileRuleLoader struct{}

func (l FileRuleLoader) LoadRule(name, env, _ string) (pkg.Resource, error) {
	path := fmt.Sprintf("./rules/%s/%s", env, name)
	return pkg.NewFileResource(path), nil
}

// DatabaseRuleLoader loads rules from database
type DatabaseRuleLoader struct {
	repo    repositories.RuleRepository
	timeout time.Duration
}

// NewDatabaseRuleLoader creates a new database rule loader
func NewDatabaseRuleLoader(repo repositories.RuleRepository) *DatabaseRuleLoader {
	return &DatabaseRuleLoader{
		repo:    repo,
		timeout: 5 * time.Second,
	}
}

func (l *DatabaseRuleLoader) LoadRule(name, env, version string) (pkg.Resource, error) {
	ctx := context.Background()

	xlog.Info(ctx, "[DATABASE_RULE_LOADER] Loading rule from database",
		xlog.String("name", name),
		xlog.String("env", env),
		xlog.String("version", version))

	var rule *repositories.Rule
	var err error

	// If version is "latest" or empty, get latest version
	if version == "latest" || version == "" {
		rule, err = l.repo.GetLatestRule(ctx, name, env)
	} else {
		rule, err = l.repo.GetRule(ctx, name, env, version)
	}

	if err != nil {
		xlog.Error(ctx, "[DATABASE_RULE_LOADER] Failed to load rule from database", xlog.Err(err))
		return nil, fmt.Errorf("failed to load rule from database: %w", err)
	}

	xlog.Info(ctx, "[DATABASE_RULE_LOADER] Rule loaded successfully from database",
		xlog.String("name", rule.Name),
		xlog.String("version", rule.Version),
		xlog.Int("size", len(rule.Content)))

	return pkg.NewBytesResource([]byte(rule.Content)), nil
}

// HybridRuleLoader tries database first, falls back to file
type HybridRuleLoader struct {
	dbLoader   *DatabaseRuleLoader
	fileLoader *FileRuleLoader
}

// NewHybridRuleLoader creates a new hybrid rule loader
func NewHybridRuleLoader(repo repositories.RuleRepository) *HybridRuleLoader {
	return &HybridRuleLoader{
		dbLoader:   NewDatabaseRuleLoader(repo),
		fileLoader: &FileRuleLoader{},
	}
}

func (l *HybridRuleLoader) LoadRule(name, env, version string) (pkg.Resource, error) {
	ctx := context.Background()

	// Try database first
	xlog.Info(ctx, "[HYBRID_RULE_LOADER] Attempting to load rule from database",
		xlog.String("name", name),
		xlog.String("env", env))

	resource, err := l.dbLoader.LoadRule(name, env, version)
	if err == nil {
		xlog.Info(ctx, "[HYBRID_RULE_LOADER] Rule loaded from database",
			xlog.String("name", name))
		return resource, nil
	}

	// Fallback to file
	xlog.Warn(ctx, "[HYBRID_RULE_LOADER] Failed to load from database, falling back to file",
		xlog.String("name", name),
		xlog.Err(err))

	resource, err = l.fileLoader.LoadRule(name, env, version)
	if err != nil {
		xlog.Error(ctx, "[HYBRID_RULE_LOADER] Failed to load from file",
			xlog.String("name", name),
			xlog.Err(err))
		return nil, fmt.Errorf("failed to load rule from both database and file: %w", err)
	}

	xlog.Info(ctx, "[HYBRID_RULE_LOADER] Rule loaded from file",
		xlog.String("name", name))

	return resource, nil
}
