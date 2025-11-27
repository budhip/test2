package transformation

import "fmt"

type TransactionStatus struct {
	value string
}

const (
	StatusSuccess = "SUCCESS"
	StatusPending = "PENDING"
	StatusFailed  = "FAILED"
)

func NewTransactionStatus(value string) (TransactionStatus, error) {
	validStatuses := map[string]bool{
		StatusSuccess: true,
		StatusPending: true,
		StatusFailed:  true,
	}

	if !validStatuses[value] {
		return TransactionStatus{}, fmt.Errorf("invalid status: %s", value)
	}

	return TransactionStatus{value: value}, nil
}

func (s TransactionStatus) String() string {
	return s.value
}

func (s TransactionStatus) IsSuccess() bool {
	return s.value == StatusSuccess
}

func (s TransactionStatus) IsPending() bool {
	return s.value == StatusPending
}

func (s TransactionStatus) IsFailed() bool {
	return s.value == StatusFailed
}

func (s TransactionStatus) ToGruleFormat() string {
	if s.IsSuccess() {
		return "1"
	}
	return "0"
}

func (s TransactionStatus) Equals(other TransactionStatus) bool {
	return s.value == other.value
}
