package papa

import (
	paymentLib "bitbucket.org/Amartha/go-payment-lib/payment-api/models"
)

type TransactionStreamEvent struct {
	paymentLib.Transaction
}

func (ts *TransactionStreamEvent) GetOTCName() string {
	if ts.PaymentMethodOTC != nil {
		return string(ts.PaymentMethodOTC.PaymentChannel)
	}
	return ""
}

func (ts *TransactionStreamEvent) GetVirtualAccountChannel() string {
	if ts.PaymentMethodVA != nil {
		return string(ts.PaymentMethodVA.PaymentChannel)
	}
	return ""
}

func (ts *TransactionStreamEvent) GetAdminFeeAmount() paymentLib.Amount {
	if adminFee, ok := ts.DetailAmount[paymentLib.DetailAmountKeyAdminFee]; ok {
		if adminFee.Type == paymentLib.DetailAmountTypeFixed {
			return adminFee.Value
		}

		if adminFee.Type == paymentLib.DetailAmountTypePercentage {
			// TODO: wait for the implementation of percentage calculation from PAPA
			// 		 currently only fixed amount is supported and used from PAPA
		}
	}
	return ""
}

func (ts *TransactionStreamEvent) GetNetAmount() paymentLib.Amount {
	if adminFee, ok := ts.DetailAmount[paymentLib.DetailAmountKeyNet]; ok {
		if adminFee.Type == paymentLib.DetailAmountTypeFixed {
			return adminFee.Value
		}
	}

	return ""
}
