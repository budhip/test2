package transformation

import "time"

type TransactionDate struct {
	value time.Time
}

func NewTransactionDate(t time.Time) TransactionDate {
	truncated := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return TransactionDate{value: truncated}
}

func (td TransactionDate) Time() time.Time {
	return td.value
}

func (td TransactionDate) Format() string {
	return td.value.Format("2006-01-02")
}

func (td TransactionDate) IsInFuture() bool {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return td.value.After(today)
}

func (td TransactionDate) IsToday() bool {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return td.value.Equal(today)
}

func (td TransactionDate) Equals(other TransactionDate) bool {
	return td.value.Equal(other.value)
}
