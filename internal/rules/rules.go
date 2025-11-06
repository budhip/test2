package rules

import (
	"context"
	"fmt"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"

	"bitbucket.org/Amartha/go-megatron/internal/config"
)

const (
	Version        = "0.0.1"
	EngineMaxCycle = 16
)

type Rule[T any] interface {
	Execute(ctx context.Context, data T) error
}

type rule struct {
	lib      *ast.KnowledgeLibrary
	ruleName string
}

var RuleLoaderVariable RuleLoader = &FileRuleLoader{}

func newRule(cfg *config.Configuration, name string, fileName string) (r rule, err error) {
	lib := ast.NewKnowledgeLibrary()
	rb := builder.NewRuleBuilder(lib)

	resource, err := RuleLoaderVariable.LoadRule(fileName, cfg.App.Env, Version)
	if err != nil {
		return r, fmt.Errorf("error loading resource rule: %w", err)
	}

	err = rb.BuildRuleFromResource(name, Version, resource)
	if err != nil {
		return r, fmt.Errorf("error build from resource: %w", err)
	}

	r.lib = lib
	r.ruleName = name

	return r, err
}

func (r rule) executeEngine(ctx context.Context, dctx ast.IDataContext) error {
	kb, err := r.lib.NewKnowledgeBaseInstance(r.ruleName, Version)
	if err != nil {
		return fmt.Errorf("error creating new knowledge base instance: %w", err)
	}
	engine := &engine.GruleEngine{
		MaxCycle: EngineMaxCycle,
	}
	matchingRules, err := engine.FetchMatchingRules(dctx, kb)
	if len(matchingRules) == 1 && matchingRules[0].RuleName == "SetDefaultTransformAcuanTransaction" {
		return errNotMatch
	}

	return engine.ExecuteWithContext(ctx, dctx, kb)
}
