package grule

import (
	"fmt"
	"regexp"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/pkg"
)

type RuleValidator struct{}

func NewRuleValidator() *RuleValidator {
	return &RuleValidator{}
}

// ValidateGruleSyntax validates if the Grule rule syntax is correct
func (v *RuleValidator) ValidateGruleSyntax(ruleContent string, ruleName string) error {
	lib := ast.NewKnowledgeLibrary()
	rb := builder.NewRuleBuilder(lib)

	resource := pkg.NewBytesResource([]byte(ruleContent))
	err := rb.BuildRuleFromResource(ruleName, "1.0.0", resource)
	if err != nil {
		return fmt.Errorf("invalid Grule syntax: %w", err)
	}

	return nil
}

// ExtractRuleNames extracts rule names from Grule content
func (v *RuleValidator) ExtractRuleNames(ruleContent string) ([]string, error) {
	// Parse and extract rule names
	// This is a simple regex-based extraction
	// You can enhance this with proper AST parsing

	pattern := regexp.MustCompile(`rule\s+(\w+)\s+`)
	matches := pattern.FindAllStringSubmatch(ruleContent, -1)

	var names []string
	for _, match := range matches {
		if len(match) >= 2 {
			names = append(names, match[1])
		}
	}

	return names, nil
}
