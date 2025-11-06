package rules

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"bitbucket.org/Amartha/go-megatron/internal/pkg/accounting"
	"github.com/hashicorp/go-multierror"
	"github.com/shopspring/decimal"
)

var (
	ProductTypeCodeGroupLoan = "1001"
	ProductTypeCodeModal     = "1011"

	AccountByProductTypeCodeDev = map[string]string{
		ProductTypeCodeGroupLoan: "121001000000012",
		ProductTypeCodeModal:     "121001000000009",
	}

	AccountByProductTypeCodeUat = map[string]string{
		ProductTypeCodeGroupLoan: "121001000000003",
		ProductTypeCodeModal:     "121001000000001",
	}

	AccountByProductTypeCodeProd = map[string]string{
		ProductTypeCodeGroupLoan: "IDR1001600011000",
		ProductTypeCodeModal:     "121001000000005",
	}
)

func (t *pasTransformed) Publish() {
	// if there are any error from transform process, skip publish
	if t.prePublishErrors != nil {
		return
	}

	t.Payload.Type = "journal_created"
	t.Payload.ProcessingDate = time.Now().In(t.timeLoc).Format(DateFormatYYYYMMDDWithTime)

	if len(t.Payload.Transactions) == 0 || len(t.Payload.Transactions)%2 != 0 {
		t.errPublish = fmt.Errorf("trx id: %s, err: transaction data not valid", t.Payload.TransactionID)
		return
	}

	for _, v := range t.Payload.Transactions {
		if v.Amount.IsZero() {
			t.errPublish = fmt.Errorf("trx id: %s, trx type: %s, err: amount is zero", t.Payload.TransactionID, v.TransactionType)
			return
		}
		if v.Account == "" {
			t.errPublish = fmt.Errorf("trx id: %s, trx type %s, err: account number is null", t.Payload.TransactionID, v.TransactionType)
			return
		}
	}
	if err := t.publisher.PublishSyncWithKeyAndLog(t.ctx, "publish transaction to journal_stream", t.cfg.Kafka.Publishers.JournalStream.Topic, t.Payload.TransactionID.String(), t.Payload); err != nil {
		t.errPublish = fmt.Errorf("trx id: %s, err: %v", t.Payload.TransactionID, err)
	}
}

func (t *pasTransformed) GetInvestedAccount(cihAccountNumber string) string {
	if cihAccountNumber == "" {
		t.prePublishErrors = multierror.Append(t.prePublishErrors, fmt.Errorf("empty account number when get invested account"))
		return ""
	}
	investedAccount, err := t.accountingClient.GetInvestedAccountNumber(t.ctx, cihAccountNumber)
	if err != nil {
		t.prePublishErrors = multierror.Append(t.prePublishErrors, err)
	}
	return investedAccount
}

func (t *pasTransformed) GetReceivableAccountNumber(cihAccountNumber string) string {
	if cihAccountNumber == "" {
		t.prePublishErrors = multierror.Append(t.prePublishErrors, fmt.Errorf("empty account number when get receivalble account"))
		return ""
	}
	receivableAccountNumber, err := t.accountingClient.GetReceivableAccountNumber(t.ctx, cihAccountNumber)
	if err != nil {
		t.prePublishErrors = multierror.Append(t.prePublishErrors, err)
	}
	return receivableAccountNumber
}

func (t *pasTransformed) GetAccountByProductTypePAS(accountNumber string) string {
	if accountNumber == "" {
		t.prePublishErrors = multierror.Append(t.prePublishErrors, fmt.Errorf("empty account number when get account"))
		return ""
	}

	account, err := t.accountingClient.GetAccountByAccountNumber(t.ctx, accountNumber)
	if err != nil {
		t.prePublishErrors = multierror.Append(t.prePublishErrors, err)
		return ""
	}

	switch t.cfg.App.Env {
	case "dev":
		return AccountByProductTypeCodeDev[account.ProductTypeCode]
	case "uat":
		return AccountByProductTypeCodeUat[account.ProductTypeCode]
	case "prod":
		return AccountByProductTypeCodeProd[account.ProductTypeCode]
	default:
		return ""
	}
}

func (t *pasTransformed) GetAccountNumberByAltID(altID string) string {
	if altID == "" {
		t.prePublishErrors = multierror.Append(t.prePublishErrors, fmt.Errorf("empty alt id when get account"))
		return ""
	}
	resp, err := t.accountingClient.GetAccountByAccountNumber(t.ctx, altID)
	if err != nil {
		t.prePublishErrors = multierror.Append(t.prePublishErrors, err)
	}

	return resp.AccountNumber
}

func (t *pasTransformed) GetByAccountNumberAltIDLegacyID(accountNumber string) string {
	if accountNumber == "" {
		t.prePublishErrors = multierror.Append(t.prePublishErrors, fmt.Errorf("empty account number when get account"))
		return ""
	}
	resp, err := t.accountingClient.GetAccountByAccountNumberOrLegacyID(t.ctx, accountNumber)
	if err != nil {
		t.prePublishErrors = multierror.Append(t.prePublishErrors, err)
	}

	return resp.AccountNumber
}

func (t *pasTransformed) GetBorrowerClaimFromAltId(altId string) string {
	input := accounting.DoGetAllAccountNumbersByParamRequest{
		AltId: altId,
	}

	resp, err := t.accountingClient.GetAccountsByParams(t.ctx, input)
	if err != nil {
		t.prePublishErrors = multierror.Append(t.prePublishErrors, err)
		return ""
	}

	key := fmt.Sprintf("%s+13102", altId)
	account, exist := resp[key]
	if !exist || len(account) == 0 {
		t.prePublishErrors = multierror.Append(t.prePublishErrors, fmt.Errorf("account not exist"))
		return ""
	}

	return account[0].AccountNumber
}

func (t *pasTransformed) TransformPoketInternalTransfer(acuan *AcuanTransaction, payload *Payload) *Payload {
	accountA := acuan.SourceAccountId
	accountB := acuan.DestinationAccountId
	getReceivableAccountA := t.GetReceivableAccountNumber(accountA)
	getReceivableAccountB := t.GetReceivableAccountNumber(accountB)

	amount := acuan.Amount
	balanceAccountA := acuan.AccountBalances[accountA].Before.ActualBalance.Value
	balanceAccountB := acuan.AccountBalances[accountB].Before.ActualBalance.Value

	transactionType := acuan.TransactionType
	transactionTypeName := acuan.TransactionTypeName
	narrative := acuan.Description

	var journals []*JournalTransaction
	if balanceAccountA.IsPositive() && (balanceAccountA.Sub(amount).IsPositive() || balanceAccountA.Sub(amount).IsZero()) {
		if balanceAccountB.IsNegative() && balanceAccountB.Add(amount).IsPositive() {
			journals = []*JournalTransaction{
				{
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             accountA,
					Narrative:           narrative,
					Amount:              balanceAccountB.Abs(),
					IsDebit:             true,
				}, {
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             getReceivableAccountB,
					Narrative:           narrative,
					Amount:              balanceAccountB.Abs(),
					IsDebit:             false,
				},
				{
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             accountA,
					Narrative:           narrative,
					Amount:              amount.Sub(balanceAccountB.Abs()),
					IsDebit:             true,
				}, {
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             accountB,
					Narrative:           narrative,
					Amount:              amount.Sub(balanceAccountB.Abs()),
					IsDebit:             false,
				},
			}
		} else if balanceAccountB.IsNegative() {
			journals = []*JournalTransaction{
				{
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             accountA,
					Narrative:           narrative,
					Amount:              amount,
					IsDebit:             true,
				}, {
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             getReceivableAccountB,
					Narrative:           narrative,
					Amount:              amount,
					IsDebit:             false,
				},
			}
		} else {
			journals = []*JournalTransaction{
				{
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             accountA,
					Narrative:           narrative,
					Amount:              amount,
					IsDebit:             true,
				}, {
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             accountB,
					Narrative:           narrative,
					Amount:              amount,
					IsDebit:             false,
				},
			}
		}
	} else if balanceAccountA.IsNegative() {
		if balanceAccountB.IsNegative() && balanceAccountB.Add(amount).IsPositive() {
			journals = []*JournalTransaction{
				{
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             getReceivableAccountA,
					Narrative:           narrative,
					Amount:              balanceAccountB.Abs(),
					IsDebit:             true,
				}, {
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             getReceivableAccountB,
					Narrative:           narrative,
					Amount:              balanceAccountB.Abs(),
					IsDebit:             false,
				},
				{
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             getReceivableAccountA,
					Narrative:           narrative,
					Amount:              amount.Sub(balanceAccountB.Abs()),
					IsDebit:             true,
				}, {
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             accountB,
					Narrative:           narrative,
					Amount:              amount.Sub(balanceAccountB.Abs()),
					IsDebit:             false,
				},
			}
		} else if balanceAccountB.IsNegative() {
			journals = []*JournalTransaction{
				{
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             getReceivableAccountA,
					Narrative:           narrative,
					Amount:              amount,
					IsDebit:             true,
				}, {
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             getReceivableAccountB,
					Narrative:           narrative,
					Amount:              amount,
					IsDebit:             false,
				},
			}
		} else {
			journals = []*JournalTransaction{
				{
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             getReceivableAccountA,
					Narrative:           narrative,
					Amount:              amount,
					IsDebit:             true,
				}, {
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             accountB,
					Narrative:           narrative,
					Amount:              amount,
					IsDebit:             false,
				},
			}
		}
	} else if balanceAccountB.IsNegative() && (balanceAccountB.Add(amount).IsNegative() || balanceAccountB.Add(amount).IsZero()) {
		if balanceAccountA.IsPositive() && balanceAccountA.Sub(amount).IsNegative() {
			journals = []*JournalTransaction{
				{
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             accountA,
					Narrative:           narrative,
					Amount:              balanceAccountA,
					IsDebit:             true,
				}, {
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             getReceivableAccountB,
					Narrative:           narrative,
					Amount:              balanceAccountA,
					IsDebit:             false,
				},
				{
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             getReceivableAccountA,
					Narrative:           narrative,
					Amount:              amount.Sub(balanceAccountA),
					IsDebit:             true,
				}, {
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             getReceivableAccountB,
					Narrative:           narrative,
					Amount:              amount.Sub(balanceAccountA),
					IsDebit:             false,
				},
			}
		} else if balanceAccountA.IsPositive() {
			journals = []*JournalTransaction{
				{
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             accountA,
					Narrative:           narrative,
					Amount:              amount,
					IsDebit:             true,
				}, {
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             getReceivableAccountB,
					Narrative:           narrative,
					Amount:              amount,
					IsDebit:             false,
				},
			}
		} else {
			journals = []*JournalTransaction{
				{
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             getReceivableAccountA,
					Narrative:           narrative,
					Amount:              amount,
					IsDebit:             true,
				}, {
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             getReceivableAccountB,
					Narrative:           narrative,
					Amount:              amount,
					IsDebit:             false,
				},
			}
		}
	} else if balanceAccountB.IsPositive() {
		if balanceAccountA.IsPositive() && balanceAccountA.Sub(amount).IsNegative() {
			journals = []*JournalTransaction{
				{
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             accountA,
					Narrative:           narrative,
					Amount:              balanceAccountA,
					IsDebit:             true,
				}, {
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             accountB,
					Narrative:           narrative,
					Amount:              balanceAccountA,
					IsDebit:             false,
				},
				{
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             getReceivableAccountA,
					Narrative:           narrative,
					Amount:              amount.Sub(balanceAccountA),
					IsDebit:             true,
				}, {
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             accountB,
					Narrative:           narrative,
					Amount:              amount.Sub(balanceAccountA),
					IsDebit:             false,
				},
			}
		} else if balanceAccountA.IsPositive() {
			journals = []*JournalTransaction{
				{
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             accountA,
					Narrative:           narrative,
					Amount:              amount,
					IsDebit:             true,
				}, {
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             accountB,
					Narrative:           narrative,
					Amount:              amount,
					IsDebit:             false,
				},
			}
		} else {
			journals = []*JournalTransaction{
				{
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             getReceivableAccountA,
					Narrative:           narrative,
					Amount:              amount,
					IsDebit:             true,
				}, {
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             accountB,
					Narrative:           narrative,
					Amount:              amount,
					IsDebit:             false,
				},
			}
		}
	} else {
		if balanceAccountA.Abs().LessThan(balanceAccountB.Abs()) {
			journals = []*JournalTransaction{
				{
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             accountA,
					Narrative:           narrative,
					Amount:              balanceAccountA,
					IsDebit:             true,
				}, {
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             getReceivableAccountB,
					Narrative:           narrative,
					Amount:              balanceAccountA,
					IsDebit:             false,
				},
				{
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             getReceivableAccountA,
					Narrative:           narrative,
					Amount:              balanceAccountA.Sub(balanceAccountB.Abs()).Abs(),
					IsDebit:             true,
				}, {
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             getReceivableAccountB,
					Narrative:           narrative,
					Amount:              balanceAccountA.Sub(balanceAccountB.Abs()).Abs(),
					IsDebit:             false,
				},
				{
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             getReceivableAccountA,
					Narrative:           narrative,
					Amount:              amount.Sub(balanceAccountB.Abs()),
					IsDebit:             true,
				}, {
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             accountB,
					Narrative:           narrative,
					Amount:              amount.Sub(balanceAccountB.Abs()),
					IsDebit:             false,
				},
			}
			if balanceAccountA.IsZero() {
				journals = []*JournalTransaction{
					{
						TransactionType:     transactionType,
						TransactionTypeName: transactionTypeName,
						Account:             getReceivableAccountA,
						Narrative:           narrative,
						Amount:              balanceAccountA.Sub(balanceAccountB.Abs()).Abs(),
						IsDebit:             true,
					}, {
						TransactionType:     transactionType,
						TransactionTypeName: transactionTypeName,
						Account:             getReceivableAccountB,
						Narrative:           narrative,
						Amount:              balanceAccountA.Sub(balanceAccountB.Abs()).Abs(),
						IsDebit:             false,
					},
					{
						TransactionType:     transactionType,
						TransactionTypeName: transactionTypeName,
						Account:             getReceivableAccountA,
						Narrative:           narrative,
						Amount:              amount.Sub(balanceAccountB.Abs()),
						IsDebit:             true,
					}, {
						TransactionType:     transactionType,
						TransactionTypeName: transactionTypeName,
						Account:             accountB,
						Narrative:           narrative,
						Amount:              amount.Sub(balanceAccountB.Abs()),
						IsDebit:             false,
					},
				}
			}
		} else if balanceAccountA.Abs().GreaterThan(balanceAccountB.Abs()) {
			journals = []*JournalTransaction{
				{
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             accountA,
					Narrative:           narrative,
					Amount:              balanceAccountB.Abs(),
					IsDebit:             true,
				}, {
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             getReceivableAccountB,
					Narrative:           narrative,
					Amount:              balanceAccountB.Abs(),
					IsDebit:             false,
				},
				{
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             accountA,
					Narrative:           narrative,
					Amount:              balanceAccountA.Sub(balanceAccountB.Abs()).Abs(),
					IsDebit:             true,
				}, {
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             accountB,
					Narrative:           narrative,
					Amount:              balanceAccountA.Sub(balanceAccountB.Abs()).Abs(),
					IsDebit:             false,
				},
				{
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             getReceivableAccountA,
					Narrative:           narrative,
					Amount:              amount.Sub(balanceAccountA).Abs(),
					IsDebit:             true,
				}, {
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             accountB,
					Narrative:           narrative,
					Amount:              amount.Sub(balanceAccountA).Abs(),
					IsDebit:             false,
				},
			}
		} else if balanceAccountA.Abs().IsZero() && balanceAccountB.Abs().IsZero() {
			journals = []*JournalTransaction{

				{
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             getReceivableAccountA,
					Narrative:           narrative,
					Amount:              amount,
					IsDebit:             true,
				}, {
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             accountB,
					Narrative:           narrative,
					Amount:              amount,
					IsDebit:             false,
				},
			}
		} else {
			journals = []*JournalTransaction{
				{
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             accountA,
					Narrative:           narrative,
					Amount:              balanceAccountA,
					IsDebit:             true,
				}, {
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             getReceivableAccountB,
					Narrative:           narrative,
					Amount:              balanceAccountA,
					IsDebit:             false,
				},
				{
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             getReceivableAccountA,
					Narrative:           narrative,
					Amount:              amount.Sub(balanceAccountA).Abs(),
					IsDebit:             true,
				}, {
					TransactionType:     transactionType,
					TransactionTypeName: transactionTypeName,
					Account:             accountB,
					Narrative:           narrative,
					Amount:              amount.Sub(balanceAccountA).Abs(),
					IsDebit:             false,
				},
			}
		}
	}
	payload.Transactions = journals

	return payload
}

func (t *pasTransformed) TransformReversalTransaction(acuan *AcuanTransaction, payload *Payload) *Payload {
	description := strings.Split(acuan.Description, " ")
	if len(description) != 2 {
		t.prePublishErrors = multierror.Append(t.prePublishErrors, errors.New("invalid description"))
		return payload
	}
	transactionIdReversal := description[1]
	transactionType := acuan.TransactionType
	transactionTypeName := acuan.TransactionTypeName
	narrative := acuan.Description

	result, err := t.accountingClient.GetJournalDetail(t.ctx, transactionIdReversal)
	if err != nil {
		t.prePublishErrors = multierror.Append(t.prePublishErrors, err)
		return payload
	}

	for _, v := range result.Journals {
		sAmount := strings.ReplaceAll(strings.ReplaceAll(v.Amount, ".", ""), ",", ".")
		amount, errDec := decimal.NewFromString(sAmount)
		if errDec != nil {
			t.prePublishErrors = multierror.Append(t.prePublishErrors, errDec)
			break
		}

		payload.Transactions = append(payload.Transactions, &JournalTransaction{
			TransactionType:     transactionType,
			TransactionTypeName: transactionTypeName,
			Account:             v.AccountNumber,
			Narrative:           narrative,
			Amount:              amount,
			IsDebit:             !v.IsDebit,
		})
	}
	return payload
}

func (t *pasTransformed) TransformReversalTransactionViaUploadAcuan(acuan *AcuanTransaction, payload *Payload) *Payload {
	transactionIdReversal := acuan.Description
	transactionType := acuan.TransactionType
	transactionTypeName := acuan.TransactionTypeName

	result, err := t.accountingClient.GetJournalDetail(t.ctx, transactionIdReversal)
	if err != nil {
		t.prePublishErrors = multierror.Append(t.prePublishErrors, err)
		return payload
	}

	for _, v := range result.Journals {
		sAmount := strings.ReplaceAll(strings.ReplaceAll(v.Amount, ".", ""), ",", ".")
		amount, errDec := decimal.NewFromString(sAmount)
		if errDec != nil {
			t.prePublishErrors = multierror.Append(t.prePublishErrors, errDec)
			break
		}

		payload.Transactions = append(payload.Transactions, &JournalTransaction{
			TransactionType:     transactionType,
			TransactionTypeName: transactionTypeName,
			Account:             v.AccountNumber,
			Narrative:           fmt.Sprintf("Reversal of %s transaction id %s", v.TransactionType, transactionIdReversal),
			Amount:              amount,
			IsDebit:             !v.IsDebit,
		})
	}
	return payload
}
