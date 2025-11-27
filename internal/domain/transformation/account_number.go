package transformation

import (
	"fmt"
	"regexp"
	"strings"
)

type AccountNumber struct {
	value string
}

func NewAccountNumber(value string) (AccountNumber, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return AccountNumber{}, nil
	}

	if len(value) < 10 || len(value) > 50 {
		return AccountNumber{}, fmt.Errorf("invalid account number length: must be 10-50 characters")
	}

	matched, _ := regexp.MatchString("^[A-Za-z0-9]+$", value)
	if !matched {
		return AccountNumber{}, fmt.Errorf("invalid account number format: must be alphanumeric")
	}

	return AccountNumber{value: value}, nil
}

func (a AccountNumber) String() string {
	return a.value
}

func (a AccountNumber) Equals(other AccountNumber) bool {
	return a.value == other.value
}

func (a AccountNumber) IsEmpty() bool {
	return a.value == ""
}

func (a AccountNumber) MaskForDisplay() string {
	if len(a.value) <= 4 {
		return strings.Repeat("X", len(a.value))
	}

	masked := strings.Repeat("X", len(a.value)-4)
	return masked + a.value[len(a.value)-4:]
}
