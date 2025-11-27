package transformation

type TransformableAmount struct {
	Amount          Amount
	TransactionType TransactionType
}

func NewTransformableAmount(amount Amount, transactionType TransactionType) TransformableAmount {
	return TransformableAmount{
		Amount:          amount,
		TransactionType: transactionType,
	}
}
