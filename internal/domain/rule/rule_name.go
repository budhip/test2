package rule

import (
	"fmt"
	"regexp"
	"strings"
)

type RuleName struct {
	value string
}

func NewRuleName(name string) (RuleName, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return RuleName{}, fmt.Errorf("rule name cannot be empty")
	}

	if len(name) < 3 || len(name) > 255 {
		return RuleName{}, fmt.Errorf("rule name must be between 3 and 255 characters")
	}

	matched, _ := regexp.MatchString("^[a-zA-Z0-9_.-]+$", name)
	if !matched {
		return RuleName{}, fmt.Errorf("rule name can only contain alphanumeric, underscore, dot, and hyphen")
	}

	return RuleName{value: name}, nil
}

func (r RuleName) String() string {
	return r.value
}

func (r RuleName) Equals(other RuleName) bool {
	return r.value == other.value
}
