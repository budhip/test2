package transformation

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type AmountBreakdown struct {
	amountType TransactionType
	amount     Amount
}

func NewAmountBreakdown(amountType string, value decimal.Decimal, currency string) (AmountBreakdown, error) {
	txType, err := NewTransactionType(amountType)
	if err != nil {
		return AmountBreakdown{}, fmt.Errorf("invalid amount type: %w", err)
	}

	amt, err := NewAmount(value, currency)
	if err != nil {
		return AmountBreakdown{}, fmt.Errorf("invalid amount: %w", err)
	}

	return AmountBreakdown{
		amountType: txType,
		amount:     amt,
	}, nil
}

func (ab AmountBreakdown) Type() TransactionType {
	return ab.amountType
}

func (ab AmountBreakdown) Amount() Amount {
	return ab.amount
}
