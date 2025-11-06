package bre

import (
	"time"

	"github.com/shopspring/decimal"
)

type (
	Event struct {
		Bill Bill `json:"bill"`
	}

	Bill struct {
		LoanID        string  `json:"loanId"`
		LoanKind      string  `json:"loanKind"`
		AccountNumber string  `json:"loanAccountNumber"`
		Bills         []Bills `json:"bills"`
	}

	Bills struct {
		ID                      int                `json:"id"`
		State                   string             `json:"state"`
		PaidAt                  time.Time          `json:"paidAt"`
		PaidPrincipalAmount     decimal.Decimal    `json:"paidPrincipalAmount"`
		PaidAmarthaMarginAmount decimal.Decimal    `json:"paidAmarthaMarginAmount"`
		ForwardRepayments       []ForwardRepayment `json:"forwardRepayments"`
	}

	ForwardRepayment struct {
		AccountNumber string          `json:"accountNumber"`
		Amount        decimal.Decimal `json:"amount"`
		Type          string          `json:"type"`
		IncomeTax     decimal.Decimal `json:"incomeTax"`
		Remark        string          `json:"remark"`
	}
)
