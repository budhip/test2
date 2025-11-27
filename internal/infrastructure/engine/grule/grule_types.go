package grule

import (
	"time"

	"github.com/shopspring/decimal"
)

type GruleAcuan struct {
	ID                       string
	Status                   string
	AccountNumber            string
	SourceAccountId          string
	DestinationAccountNumber string
	RefNumber                string
	TransactionType          string
	TransactionTime          time.Time
	TransactionFlow          string
	NetAmount                decimal.Decimal
	Amount                   decimal.Decimal
	Currency                 string
	Description              string
	Metadata                 map[string]interface{}
	CreatedAt                time.Time
}

type GruleJournal struct {
	Transactions *GruleTransactionList
}

type GruleTransaction struct {
	IsReadyToPublish bool
}

type GruleTransactionList struct {
	Items []GruleJournalEntry
}

func (t *GruleTransactionList) Append(entry GruleJournalEntry) {
	t.Items = append(t.Items, entry)
}

type GruleJournalEntry struct {
	Account         string
	Narrative       string
	Amount          decimal.Decimal
	TransactionDate string
	Status          string
	TypeTransaction string
	OrderType       string
	RefNumber       string
	Description     string
	Metadata        map[string]interface{}
	OrderTime       time.Time
	TransactionTime time.Time
	Currency        string
	TransactionID   string
}

type GruleJournalDebitCredit struct {
	Account         string
	Narrative       string
	Amount          decimal.Decimal
	TransactionDate string
	Status          string
	TypeTransaction string
	OrderType       string
	RefNumber       string
	Description     string
	Metadata        map[string]interface{}
	OrderTime       time.Time
	TransactionTime time.Time
	Currency        string
	TransactionID   string
}
