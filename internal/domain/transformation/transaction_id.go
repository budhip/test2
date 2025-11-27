package transformation

import (
	"strings"

	"github.com/google/uuid"
)

type TransactionID struct {
	value string
}

func NewTransactionID(id string) (TransactionID, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		id = uuid.New().String()
	}

	return TransactionID{value: id}, nil
}

func GenerateTransactionID() TransactionID {
	return TransactionID{value: uuid.New().String()}
}

func (t TransactionID) String() string {
	return t.value
}

func (t TransactionID) Equals(other TransactionID) bool {
	return t.value == other.value
}
