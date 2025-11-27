package transformation

import (
	"fmt"
	"strings"
)

type WalletTransactionID struct {
	value string
}

func NewWalletTransactionID(id string) (WalletTransactionID, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return WalletTransactionID{}, fmt.Errorf("wallet transaction ID cannot be empty")
	}

	if len(id) < 5 || len(id) > 100 {
		return WalletTransactionID{}, fmt.Errorf("wallet transaction ID must be between 5 and 100 characters")
	}

	return WalletTransactionID{value: id}, nil
}

func (w WalletTransactionID) String() string {
	return w.value
}

func (w WalletTransactionID) Equals(other WalletTransactionID) bool {
	return w.value == other.value
}

func (w WalletTransactionID) IsEmpty() bool {
	return w.value == ""
}
