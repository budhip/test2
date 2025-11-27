package transformation

import "errors"

var (
	ErrInvalidStatusForTransformation = errors.New("transaction status must be SUCCESS for transformation")
	ErrMissingAccountNumber           = errors.New("account number is required")
	ErrNoAmountToTransform            = errors.New("no amount to transform")

	ErrSameFromToAccount     = errors.New("from and to accounts cannot be the same")
	ErrNegativeAmount        = errors.New("amount cannot be negative")
	ErrFutureTransactionDate = errors.New("transaction date cannot be in future")

	ErrRuleNotFound         = errors.New("transformation rule not found")
	ErrTransformationFailed = errors.New("transformation failed")
)

type DomainEvent interface {
	EventType() string
	OccurredAt() interface{}
}

type WalletTransactionTransformedEvent struct {
	walletTransactionID WalletTransactionID
	occurredAt          interface{}
}

func NewWalletTransactionTransformedEvent(id WalletTransactionID, occurredAt interface{}) DomainEvent {
	return &WalletTransactionTransformedEvent{
		walletTransactionID: id,
		occurredAt:          occurredAt,
	}
}

func (e *WalletTransactionTransformedEvent) EventType() string {
	return "WalletTransactionTransformed"
}

func (e *WalletTransactionTransformedEvent) OccurredAt() interface{} {
	return e.occurredAt
}
