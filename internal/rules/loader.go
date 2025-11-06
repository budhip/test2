package rules

import (
	"fmt"

	"github.com/hyperjumptech/grule-rule-engine/pkg"
)

// RuleLoader is interface to get rule resource, it can be accessed from local files or sql database
type RuleLoader interface {
	LoadRule(name, env, version string) (pkg.Resource, error)
}

type FileRuleLoader struct{}

func (l FileRuleLoader) LoadRule(name, env, _ string) (pkg.Resource, error) {
	path := fmt.Sprintf("./rules/%s/%s", env, name)
	return pkg.NewFileResource(path), nil
}
