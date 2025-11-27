package transformation

import "fmt"

type Currency struct {
	code string
}

func NewCurrency(code string) (Currency, error) {
	validCurrencies := map[string]bool{
		"IDR": true,
		"USD": true,
		"SGD": true,
		"EUR": true,
		"MYR": true,
		"JPY": true,
		"GBP": true,
		"CNY": true,
		"THB": true,
	}

	if !validCurrencies[code] {
		return Currency{}, fmt.Errorf("invalid currency code: %s", code)
	}

	return Currency{code: code}, nil
}

func (c Currency) Code() string {
	return c.code
}

func (c Currency) Equals(other Currency) bool {
	return c.code == other.code
}

func (c Currency) String() string {
	return c.code
}
