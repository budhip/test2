package transformation

import "fmt"

type TransformationValidator struct{}

func NewTransformationValidator() *TransformationValidator {
	return &TransformationValidator{}
}

func (v *TransformationValidator) ValidateWalletTransaction(wt *WalletTransaction) error {
	if err := wt.ValidateForTransformation(); err != nil {
		return err
	}

	if wt.Status().IsSuccess() {
		if wt.TransactionTime().IsZero() {
			return fmt.Errorf("success transaction must have transaction time")
		}
	}

	if !wt.NetAmount().IsZero() {
		for _, breakdown := range wt.Amounts() {
			if !breakdown.Amount().Currency().Equals(wt.NetAmount().Currency()) {
				return fmt.Errorf("currency mismatch: net amount %s vs breakdown %s",
					wt.NetAmount().Currency().Code(),
					breakdown.Amount().Currency().Code())
			}
		}
	}

	return nil
}

func (v *TransformationValidator) ValidateTransaction(tx *Transaction) error {
	return tx.Validate()
}
