package models

import (
	"github.com/shopspring/decimal"
	"time"
)

// Acuan represents the input wallet transaction for Grule
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
	Amount                   decimal.Decimal // Current amount being processed
	Currency                 string
	Description              string
	Metadata                 map[string]interface{}
	CreatedAt                time.Time
}

// GruleAcuanJournal represents the output journal/transaction collection
type GruleAcuanJournal struct {
	Transactions *GruleAcuanTransactionList
}

// GruleAcuanTransaction controls the transformation flow
type GruleAcuanTransaction struct {
	IsReadyToPublish bool
}

// GruleAcuanTransactionList holds the list of transformed transactions
type GruleAcuanTransactionList struct {
	Items []GruleAcuanJournalEntry
}

func (t *GruleAcuanTransactionList) Append(entry GruleAcuanJournalEntry) {
	t.Items = append(t.Items, entry)
}

// GruleAcuanJournalEntry represents a single debit/credit entry
type GruleAcuanJournalEntry struct {
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

// Helper objects for rule writing
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
