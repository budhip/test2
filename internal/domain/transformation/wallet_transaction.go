package transformation

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type WalletTransaction struct {
	id                       WalletTransactionID
	status                   TransactionStatus
	accountNumber            AccountNumber
	destinationAccountNumber AccountNumber
	refNumber                RefNumber
	transactionType          TransactionType
	transactionTime          time.Time
	transactionFlow          TransactionFlow
	netAmount                Amount
	amounts                  []AmountBreakdown
	description              string
	metadata                 Metadata
	createdAt                time.Time
	events                   []DomainEvent
}

func NewWalletTransaction(
	id string,
	status string,
	accountNumber string,
	destinationAccountNumber string,
	refNumber string,
	transactionType string,
	transactionTime time.Time,
	transactionFlow string,
	netAmount decimal.Decimal,
	currency string,
	description string,
	metadata map[string]interface{},
) (*WalletTransaction, error) {

	walletTxID, err := NewWalletTransactionID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID: %w", err)
	}

	txStatus, err := NewTransactionStatus(status)
	if err != nil {
		return nil, fmt.Errorf("invalid status: %w", err)
	}

	srcAccount, err := NewAccountNumber(accountNumber)
	if err != nil {
		return nil, fmt.Errorf("invalid account number: %w", err)
	}

	destAccount, err := NewAccountNumber(destinationAccountNumber)
	if err != nil {
		return nil, fmt.Errorf("invalid destination account: %w", err)
	}

	ref, err := NewRefNumber(refNumber)
	if err != nil {
		return nil, fmt.Errorf("invalid ref number: %w", err)
	}

	txType, err := NewTransactionType(transactionType)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction type: %w", err)
	}

	flow, err := NewTransactionFlow(transactionFlow)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction flow: %w", err)
	}

	amount, err := NewAmount(netAmount, currency)
	if err != nil {
		return nil, fmt.Errorf("invalid net amount: %w", err)
	}

	meta := NewMetadata(metadata)

	return &WalletTransaction{
		id:                       walletTxID,
		status:                   txStatus,
		accountNumber:            srcAccount,
		destinationAccountNumber: destAccount,
		refNumber:                ref,
		transactionType:          txType,
		transactionTime:          transactionTime,
		transactionFlow:          flow,
		netAmount:                amount,
		amounts:                  []AmountBreakdown{},
		description:              description,
		metadata:                 meta,
		createdAt:                time.Now(),
		events:                   []DomainEvent{},
	}, nil
}

func (wt *WalletTransaction) AddAmountBreakdown(amountType string, value decimal.Decimal, currency string) error {
	for _, existing := range wt.amounts {
		if existing.Type().String() == amountType {
			return fmt.Errorf("amount breakdown type already exists: %s", amountType)
		}
	}

	breakdown, err := NewAmountBreakdown(amountType, value, currency)
	if err != nil {
		return err
	}

	wt.amounts = append(wt.amounts, breakdown)
	return nil
}

func (wt *WalletTransaction) ValidateForTransformation() error {
	if !wt.status.IsSuccess() {
		return ErrInvalidStatusForTransformation
	}

	if wt.accountNumber.IsEmpty() {
		return ErrMissingAccountNumber
	}

	if wt.netAmount.IsZero() && len(wt.amounts) == 0 {
		return ErrNoAmountToTransform
	}

	return nil
}

func (wt *WalletTransaction) GetAmountsForTransformation() []TransformableAmount {
	var result []TransformableAmount

	if !wt.netAmount.IsZero() {
		result = append(result, TransformableAmount{
			Amount:          wt.netAmount,
			TransactionType: wt.transactionType,
		})
	}

	for _, breakdown := range wt.amounts {
		if !breakdown.Amount().IsZero() {
			result = append(result, TransformableAmount{
				Amount:          breakdown.Amount(),
				TransactionType: breakdown.Type(),
			})
		}
	}

	return result
}

func (wt *WalletTransaction) MarkAsTransformed() {
	event := NewWalletTransactionTransformedEvent(wt.id, time.Now())
	wt.events = append(wt.events, event)
}

func (wt *WalletTransaction) ID() WalletTransactionID {
	return wt.id
}

func (wt *WalletTransaction) Status() TransactionStatus {
	return wt.status
}

func (wt *WalletTransaction) AccountNumber() AccountNumber {
	return wt.accountNumber
}

func (wt *WalletTransaction) DestinationAccountNumber() AccountNumber {
	return wt.destinationAccountNumber
}

func (wt *WalletTransaction) RefNumber() RefNumber {
	return wt.refNumber
}

func (wt *WalletTransaction) TransactionType() TransactionType {
	return wt.transactionType
}

func (wt *WalletTransaction) TransactionTime() time.Time {
	return wt.transactionTime
}

func (wt *WalletTransaction) TransactionFlow() TransactionFlow {
	return wt.transactionFlow
}

func (wt *WalletTransaction) NetAmount() Amount {
	return wt.netAmount
}

func (wt *WalletTransaction) Amounts() []AmountBreakdown {
	return wt.amounts
}

func (wt *WalletTransaction) Description() string {
	return wt.description
}

func (wt *WalletTransaction) Metadata() Metadata {
	return wt.metadata
}

func (wt *WalletTransaction) CreatedAt() time.Time {
	return wt.createdAt
}

func (wt *WalletTransaction) DomainEvents() []DomainEvent {
	return wt.events
}
