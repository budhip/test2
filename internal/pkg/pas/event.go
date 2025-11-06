package pas

import (
	goAcuanLibModel "bitbucket.org/Amartha/go-acuan-lib/model"
	"github.com/shopspring/decimal"
)

type (
	OutMessage struct {
		Identifier string                                             `json:"identifier"`
		Status     string                                             `json:"status"`
		AcuanData  goAcuanLibModel.Payload[goAcuanLibModel.DataOrder] `json:"acuanData"`
		Message    string                                             `json:"message"`

		AccountBalances   map[string]AccountBalance `json:"accountBalances"`
		WalletTransaction WalletTransaction         `json:"walletTransaction"`
		IsRetry           bool                      `json:"isRetry"`
	}

	AccountBalance struct {
		Before Balance `json:"before"`
		After  Balance `json:"after"`
	}
	Balance struct {
		ActualBalance ValueCurrency `json:"actualBalance"`
		LastUpdatedAt string        `json:"lastUpdatedAt"`
		Version       int           `json:"version"`
	}

	WalletTransaction struct {
		AccountNumber            string        `json:"accountNumber"`
		DestinationAccountNumber string        `json:"destinationAccountNumber"`
		NetAmount                ValueCurrency `json:"netAmount"`
		Amounts                  []Amount      `json:"amounts"`
	}
	Amount struct {
		Type   string        `json:"type"`
		Amount ValueCurrency `json:"amount"`
	}

	ValueCurrency struct {
		Value    decimal.Decimal `json:"value"`
		Currency string          `json:"currency"`
	}
)
