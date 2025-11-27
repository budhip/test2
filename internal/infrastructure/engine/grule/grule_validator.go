package grule

import (
	"fmt"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/pkg"
)

type GruleValidator struct{}

func NewGruleValidator() *GruleValidator {
	return &GruleValidator{}
}

func (v *GruleValidator) ValidateGruleSyntax(content string, ruleName string) error {
	lib := ast.NewKnowledgeLibrary()
	rb := builder.NewRuleBuilder(lib)

	resource := pkg.NewBytesResource([]byte(content))
	err := rb.BuildRuleFromResource(ruleName, "1.0.0", resource)
	if err != nil {
		return fmt.Errorf("invalid Grule syntax: %w", err)
	}

	return nil
}
