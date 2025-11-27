package rule

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type RuleID struct {
	value string
}

func NewRuleID() RuleID {
	return RuleID{value: uuid.New().String()}
}

func ParseRuleID(id string) (RuleID, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return RuleID{}, fmt.Errorf("rule ID cannot be empty")
	}

	if _, err := uuid.Parse(id); err != nil {
		return RuleID{}, fmt.Errorf("invalid UUID format: %w", err)
	}

	return RuleID{value: id}, nil
}

func (r RuleID) String() string {
	return r.value
}

func (r RuleID) Equals(other RuleID) bool {
	return r.value == other.value
}

func (r RuleID) IsEmpty() bool {
	return r.value == ""
}
