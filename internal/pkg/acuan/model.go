package acuan

import (
	"time"

	goAcuanLibModel "bitbucket.org/Amartha/go-acuan-lib/model"
	"github.com/shopspring/decimal"
)

type PublishOrderRequest struct {
	OrderType    string
	RefNumber    string
	Transactions []OrderTransaction
}

type OrderTransaction struct {
	FromAccount     string
	ToAccount       string
	Amount          decimal.Decimal
	Method          string
	TransactionType string
	TransactionTime time.Time
	Description     string
	Metadata        interface{}
	Currency        string
}

func (e OrderTransaction) ToAcuanTransaction() goAcuanLibModel.Transaction {
	currency := goAcuanLibModel.TransactionCurrencyIDR
	if e.Currency != "" {
		currency = e.Currency
	}

	return goAcuanLibModel.Transaction{
		Amount:               e.Amount,
		Currency:             currency,
		SourceAccountId:      e.FromAccount,
		DestinationAccountId: e.ToAccount,
		Description:          e.Description,
		Method:               goAcuanLibModel.TransactionMethod(e.Method),
		TransactionType:      goAcuanLibModel.TransactionType(e.TransactionType),
		TransactionTime:      goAcuanLibModel.AcuanTime{Time: &e.TransactionTime},
		Status:               goAcuanLibModel.TransactionStatusSuccess,
		Meta:                 e.Metadata,
	}
}
