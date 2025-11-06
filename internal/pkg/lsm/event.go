package lsm

import (
	"time"

	"github.com/shopspring/decimal"
)

type (
	Event struct {
		Loan Loan `json:"loan"`
	}

	Loan struct {
		ID                string    `json:"id"`
		DisbursedDate     time.Time `json:"disbursedDate"`
		AccountNumber     string    `json:"accountNumber"`
		LoanAccountNumber string    `json:"loanAccountNumber"`
		Principal         Amount    `json:"principal"`
		Tenor             Tenor     `json:"tenor"`
		Fee               []Fee     `json:"fee"`
		State             string    `json:"state"`
		SubState          string    `json:"subState"`
		Kind              string    `json:"kind"`
		Version           int64     `json:"version,omitempty"`
	}

	Amount struct {
		Amount   decimal.Decimal `json:"amount"`
		Currency string          `json:"currency"`
	}

	Tenor struct {
		Unit  string `json:"unit"`
		Value uint64 `json:"value"`
	}

	Fee struct {
		Kind      string          `json:"kind"`
		Value     decimal.Decimal `json:"value"`
		Method    string          `json:"method"`
		ValueKind string          `json:"valueKind"`
	}
)

func (e Loan) GetAdminFee() *Fee {
	for _, fee := range e.Fee {
		if fee.Kind == "admin" {
			return &fee
		}
	}

	return nil
}

func (t Tenor) GetByMonth() float64 {
	const (
		daysInMonth  float64 = 30
		daysInWeek   float64 = 7
		monthsInYear float64 = 12
	)
	switch t.Unit {
	case "DAILY":
		return float64(t.Value) / daysInMonth
	case "WEEKLY":
		return float64(t.Value) * (daysInWeek / daysInMonth)
	case "MONTHLY":
		return float64(t.Value)
	case "YEARLY":
		return float64(t.Value) * monthsInYear
	}

	return 0
}
