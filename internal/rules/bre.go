package rules

import (
	"context"
	"time"

	"github.com/hashicorp/go-multierror"

	"bitbucket.org/Amartha/go-megatron/internal/config"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/accounting"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/acuan"
	brePkg "bitbucket.org/Amartha/go-megatron/internal/pkg/bre"
	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/shopspring/decimal"
)

const (
	BRERulename     = "BRE Rule"
	BRERuleFilename = "bre.grl"
)

type bre struct {
	rule
	acuanClient      acuan.AcuanClient
	accountingClient accounting.Client
}

type acuanBRETransaction struct {
	publisher
	acuan.PublishOrderRequest

	accountingClient accounting.Client

	// prePublishErrors is errors that happened before publishing
	// usually it's a validation error or unable to get data from external service
	prePublishErrors *multierror.Error

	PPNAmount            decimal.Decimal
	AmarthaRevenueAmount decimal.Decimal

	A           decimal.Decimal // For calculating PPN (11%)
	B           decimal.Decimal // For calculating PPN (1.11)
	RoundPlaces int32           // For rounding

	PPNPartnerAccount     string
	AmarthaRevenueAccount string
	PPHLenderAccount      string
	SystemAccount         string

	IsReadyToPublish              bool
	IsForwardRepaymentProcessDone bool
	IsSetAccountsDone             bool
	Now                           string
}

func (t *acuanBRETransaction) AddTransaction(
	fromAccount,
	toAccount,
	method,
	transactionType,
	description string,
	transactionTime time.Time,
	amount decimal.Decimal,
) {
	if !amount.IsZero() {
		t.Transactions = append(t.Transactions, acuan.OrderTransaction{
			FromAccount:     fromAccount,
			ToAccount:       toAccount,
			Amount:          amount,
			Method:          method,
			TransactionType: transactionType,
			TransactionTime: transactionTime,
			Description:     description,
			Metadata: map[string]string{
				"description": description,
			},
			Currency: "IDR",
		})
	}
}

func (t *acuanBRETransaction) GetInvestedAccount(cihAccountNumber string) string {
	investedAccount, err := t.accountingClient.GetInvestedAccountNumber(t.ctx, cihAccountNumber)
	if err != nil {
		t.prePublishErrors = multierror.Append(t.prePublishErrors, err)
	}

	return investedAccount
}

func (t *acuanBRETransaction) Publish() {
	t.errPublish = t.publisher.acuanClient.PublishTransaction(t.publisher.ctx, t.PublishOrderRequest)
}

func NewBre(
	cfg *config.Configuration,
	acuanClient acuan.AcuanClient,
	accountingClient accounting.Client,
) (Rule[brePkg.Event], error) {
	r, err := newRule(cfg, BRERulename, BRERuleFilename)

	b := &bre{
		rule:             r,
		acuanClient:      acuanClient,
		accountingClient: accountingClient,
	}

	return b, err
}

func (r bre) Execute(ctx context.Context, data brePkg.Event) error {
	bill := data.Bill

	for _, b := range bill.Bills {
		tx := &acuanBRETransaction{
			accountingClient: r.accountingClient,
			publisher: publisher{
				acuanClient: r.acuanClient,
				ctx:         ctx,
			},
			PPNPartnerAccount:     "0",
			AmarthaRevenueAccount: "0",
			A:                     decimal.NewFromFloat(0.11),
			B:                     decimal.NewFromFloat(1.11),
			Now:                   time.Now().Format("20060102"),
		}
		// Set Forward Repayment related transactions
		for _, fr := range b.ForwardRepayments {
			dctx := ast.NewDataContext()

			if fr.Amount.IsZero() {
				continue
			}

			if err := dctx.Add("Bill", &bill); err != nil {
				return err
			}

			if err := dctx.Add("Bills", &b); err != nil {
				return err
			}

			if err := dctx.Add("ForwardRepayment", &fr); err != nil {
				return err
			}

			if err := dctx.Add("Transaction", tx); err != nil {
				return err
			}

			if err := r.executeEngine(ctx, dctx); err != nil {
				xlog.Info(ctx, "failed to execute with context", xlog.Err(err))
				return err
			}
		}

		// whenether forward repayment populating process done, publish
		tx.IsForwardRepaymentProcessDone = true

		dctx := ast.NewDataContext()
		fr := brePkg.ForwardRepayment{}

		if err := dctx.Add("Bill", &bill); err != nil {
			return err
		}

		if err := dctx.Add("Bills", &b); err != nil {
			return err
		}

		if err := dctx.Add("Transaction", tx); err != nil {
			return err
		}

		if err := dctx.Add("ForwardRepayment", &fr); err != nil {
			return err
		}

		if err := r.executeEngine(ctx, dctx); err != nil {
			xlog.Warn(ctx, "failed to execute with context", xlog.Err(err))
			return err
		}

		prePublishErrors := tx.prePublishErrors.ErrorOrNil()
		if prePublishErrors != nil {
			xlog.Warn(ctx, "failed to execute rule", xlog.Err(prePublishErrors))
			return tx.prePublishErrors
		}

		if tx.errPublish != nil {
			xlog.Warn(ctx, "failed to publish", xlog.Err(tx.errPublish))
			return tx.errPublish
		}
	}

	return nil
}
