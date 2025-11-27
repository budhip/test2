package transformation

import "fmt"

type TransactionType struct {
	value string
}

const (
	TypeTopUp            = "TOPUP"
	TypeCashout          = "CASHOUT"
	TypeInvestment       = "INVESTMENT"
	TypeDisbursement     = "DISBURSEMENT"
	TypeRepayment        = "REPAYMENT"
	TypeRefund           = "REFUND"
	TypeAdminFee         = "ADMIN_FEE"
	TypeInsurancePremium = "INSURANCE_PREMIUM"
)

func NewTransactionType(value string) (TransactionType, error) {
	validTypes := map[string]bool{
		TypeTopUp:            true,
		TypeCashout:          true,
		TypeInvestment:       true,
		TypeDisbursement:     true,
		TypeRepayment:        true,
		TypeRefund:           true,
		TypeAdminFee:         true,
		TypeInsurancePremium: true,
	}

	if !validTypes[value] {
		return TransactionType{}, fmt.Errorf("invalid transaction type: %s", value)
	}

	return TransactionType{value: value}, nil
}

func (t TransactionType) String() string {
	return t.value
}

func (t TransactionType) Equals(other TransactionType) bool {
	return t.value == other.value
}

func (t TransactionType) IsDebit() bool {
	debitTypes := []string{TypeCashout, TypeInvestment, TypeAdminFee, TypeInsurancePremium}
	for _, dt := range debitTypes {
		if t.value == dt {
			return true
		}
	}
	return false
}

func (t TransactionType) IsCredit() bool {
	return !t.IsDebit()
}

func (t TransactionType) RequiresApproval() bool {
	return t.value == TypeCashout || t.value == TypeInvestment
}
