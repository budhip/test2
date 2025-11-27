package transformation

import (
	"fmt"
	"strings"
)

type RefNumber struct {
	value string
}

func NewRefNumber(value string) (RefNumber, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return RefNumber{}, fmt.Errorf("ref number cannot be empty")
	}

	if len(value) < 5 || len(value) > 100 {
		return RefNumber{}, fmt.Errorf("ref number must be between 5 and 100 characters")
	}

	return RefNumber{value: value}, nil
}

func (r RefNumber) String() string {
	return r.value
}

func (r RefNumber) Equals(other RefNumber) bool {
	return r.value == other.value
}
