package postgres

import (
	"time"

	"github.com/shopspring/decimal"
)

type DBRule struct {
	ID        string
	Name      string
	Env       string
	Version   string
	Content   string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy string
	UpdatedBy string
}

type DBWalletTransaction struct {
	ID                       string
	Status                   string
	AccountNumber            string
	DestinationAccountNumber string
	RefNumber                string
	TransactionType          string
	TransactionTime          time.Time
	TransactionFlow          string
	NetAmount                decimal.Decimal
	Currency                 string
	Description              string
	Metadata                 []byte
	CreatedAt                time.Time
}

type DBTransaction struct {
	ID              string
	FromAccount     string
	ToAccount       string
	FromNarrative   string
	ToNarrative     string
	TransactionDate time.Time
	Amount          decimal.Decimal
	Currency        string
	Status          string
	TransactionType string
	Description     string
	RefNumber       string
	OrderType       string
	OrderTime       time.Time
	TransactionTime time.Time
	Metadata        []byte
	CreatedAt       time.Time
}
