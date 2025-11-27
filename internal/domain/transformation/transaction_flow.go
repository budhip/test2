package transformation

import "fmt"

type TransactionFlow struct {
	value string
}

const (
	FlowDebit  = "DEBIT"
	FlowCredit = "CREDIT"
)

func NewTransactionFlow(flow string) (TransactionFlow, error) {
	validFlows := map[string]bool{
		FlowDebit:  true,
		FlowCredit: true,
	}

	if !validFlows[flow] {
		return TransactionFlow{}, fmt.Errorf("invalid transaction flow: must be DEBIT or CREDIT")
	}

	return TransactionFlow{value: flow}, nil
}

func (t TransactionFlow) String() string {
	return t.value
}

func (t TransactionFlow) IsDebit() bool {
	return t.value == FlowDebit
}

func (t TransactionFlow) IsCredit() bool {
	return t.value == FlowCredit
}

func (t TransactionFlow) Equals(other TransactionFlow) bool {
	return t.value == other.value
}
