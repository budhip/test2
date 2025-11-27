package transformation

import (
	"fmt"

	"github.com/shopspring/decimal"
)

type Amount struct {
	value    decimal.Decimal
	currency Currency
}

func NewAmount(value decimal.Decimal, currencyCode string) (Amount, error) {
	if value.IsNegative() {
		return Amount{}, fmt.Errorf("amount cannot be negative: %s", value.String())
	}

	currency, err := NewCurrency(currencyCode)
	if err != nil {
		return Amount{}, err
	}

	return Amount{
		value:    value,
		currency: currency,
	}, nil
}

func (a Amount) Value() decimal.Decimal {
	return a.value
}

func (a Amount) Currency() Currency {
	return a.currency
}

func (a Amount) IsZero() bool {
	return a.value.IsZero()
}

func (a Amount) IsNegative() bool {
	return a.value.IsNegative()
}

func (a Amount) IsPositive() bool {
	return a.value.IsPositive()
}

func (a Amount) Add(other Amount) (Amount, error) {
	if !a.currency.Equals(other.currency) {
		return Amount{}, fmt.Errorf("cannot add different currencies: %s and %s",
			a.currency.Code(), other.currency.Code())
	}

	return Amount{
		value:    a.value.Add(other.value),
		currency: a.currency,
	}, nil
}

func (a Amount) Subtract(other Amount) (Amount, error) {
	if !a.currency.Equals(other.currency) {
		return Amount{}, fmt.Errorf("cannot subtract different currencies")
	}

	return Amount{
		value:    a.value.Sub(other.value),
		currency: a.currency,
	}, nil
}

func (a Amount) Multiply(multiplier decimal.Decimal) Amount {
	return Amount{
		value:    a.value.Mul(multiplier),
		currency: a.currency,
	}
}

func (a Amount) String() string {
	return fmt.Sprintf("%s %s", a.currency.Code(), a.value.String())
}

func (a Amount) Equals(other Amount) bool {
	return a.value.Equal(other.value) && a.currency.Equals(other.currency)
}
