package transformation

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type Transaction struct {
	id              TransactionID
	fromAccount     AccountNumber
	toAccount       AccountNumber
	fromNarrative   string
	toNarrative     string
	transactionDate TransactionDate
	amount          Amount
	status          TransactionStatus
	method          string
	transactionType TransactionType
	description     string
	refNumber       RefNumber
	orderType       OrderType
	orderTime       time.Time
	transactionTime time.Time
	metadata        Metadata
	createdAt       time.Time
}

func NewTransaction(
	id string,
	fromAccount string,
	toAccount string,
	transactionDate time.Time,
	amount decimal.Decimal,
	currency string,
	status string,
	transactionType string,
	description string,
	refNumber string,
	orderType string,
	orderTime time.Time,
	transactionTime time.Time,
	metadata map[string]interface{},
) (*Transaction, error) {

	txID, err := NewTransactionID(id)
	if err != nil {
		return nil, err
	}

	from, err := NewAccountNumber(fromAccount)
	if err != nil {
		return nil, fmt.Errorf("invalid from account: %w", err)
	}

	to, err := NewAccountNumber(toAccount)
	if err != nil {
		return nil, fmt.Errorf("invalid to account: %w", err)
	}

	txDate := NewTransactionDate(transactionDate)

	amt, err := NewAmount(amount, currency)
	if err != nil {
		return nil, err
	}

	txStatus, err := NewTransactionStatus(status)
	if err != nil {
		return nil, err
	}

	txType, err := NewTransactionType(transactionType)
	if err != nil {
		return nil, err
	}

	ref, err := NewRefNumber(refNumber)
	if err != nil {
		return nil, err
	}

	ordType, err := NewOrderType(orderType)
	if err != nil {
		return nil, err
	}

	meta := NewMetadata(metadata)

	return &Transaction{
		id:              txID,
		fromAccount:     from,
		toAccount:       to,
		transactionDate: txDate,
		amount:          amt,
		status:          txStatus,
		transactionType: txType,
		description:     description,
		refNumber:       ref,
		orderType:       ordType,
		orderTime:       orderTime,
		transactionTime: transactionTime,
		metadata:        meta,
		createdAt:       time.Now(),
	}, nil
}

func (t *Transaction) Validate() error {
	if t.fromAccount.Equals(t.toAccount) {
		return ErrSameFromToAccount
	}

	if t.amount.IsNegative() {
		return ErrNegativeAmount
	}

	if t.transactionDate.IsInFuture() {
		return ErrFutureTransactionDate
	}

	return nil
}

func (t *Transaction) SetNarratives(fromNarrative, toNarrative string) {
	t.fromNarrative = fromNarrative
	t.toNarrative = toNarrative
}

func (t *Transaction) ID() TransactionID {
	return t.id
}

func (t *Transaction) FromAccount() AccountNumber {
	return t.fromAccount
}

func (t *Transaction) ToAccount() AccountNumber {
	return t.toAccount
}

func (t *Transaction) FromNarrative() string {
	return t.fromNarrative
}

func (t *Transaction) ToNarrative() string {
	return t.toNarrative
}

func (t *Transaction) TransactionDate() TransactionDate {
	return t.transactionDate
}

func (t *Transaction) Amount() Amount {
	return t.amount
}

func (t *Transaction) Status() TransactionStatus {
	return t.status
}

func (t *Transaction) Method() string {
	return t.method
}

func (t *Transaction) TransactionType() TransactionType {
	return t.transactionType
}

func (t *Transaction) Description() string {
	return t.description
}

func (t *Transaction) RefNumber() RefNumber {
	return t.refNumber
}

func (t *Transaction) OrderType() OrderType {
	return t.orderType
}

func (t *Transaction) OrderTime() time.Time {
	return t.orderTime
}

func (t *Transaction) TransactionTime() time.Time {
	return t.transactionTime
}

func (t *Transaction) Metadata() Metadata {
	return t.metadata
}

func (t *Transaction) CreatedAt() time.Time {
	return t.createdAt
}
