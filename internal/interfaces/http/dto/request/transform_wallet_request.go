package request

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type TransformWalletTransactionRequest struct {
	WalletTransaction WalletTransactionInput `json:"walletTransaction"`
}

type WalletTransactionInput struct {
	ID                       string                 `json:"id"`
	Status                   string                 `json:"status"`
	AccountNumber            string                 `json:"accountNumber"`
	DestinationAccountNumber string                 `json:"destinationAccountNumber"`
	RefNumber                string                 `json:"refNumber"`
	TransactionType          string                 `json:"transactionType"`
	TransactionTime          time.Time              `json:"transactionTime"`
	TransactionFlow          string                 `json:"transactionFlow"`
	NetAmount                AmountInput            `json:"netAmount"`
	Amounts                  []AmountBreakdownInput `json:"amounts"`
	Description              string                 `json:"description"`
	Metadata                 map[string]interface{} `json:"metadata"`
	CreatedAt                time.Time              `json:"createdAt"`
}

type AmountInput struct {
	Value    decimal.Decimal `json:"value"`
	Currency string          `json:"currency"`
}

type AmountBreakdownInput struct {
	Type   string      `json:"type"`
	Amount AmountInput `json:"amount"`
}

func (r *TransformWalletTransactionRequest) Validate() error {
	if r.WalletTransaction.ID == "" {
		return fmt.Errorf("wallet transaction ID is required")
	}

	if r.WalletTransaction.TransactionType == "" {
		return fmt.Errorf("transaction type is required")
	}

	if r.WalletTransaction.AccountNumber == "" {
		return fmt.Errorf("account number is required")
	}

	if r.WalletTransaction.NetAmount.Value.IsZero() && len(r.WalletTransaction.Amounts) == 0 {
		return fmt.Errorf("either net amount or breakdown amounts must be provided")
	}

	return nil
}
