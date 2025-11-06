package rules

import (
	"context"
	"time"

	"bitbucket.org/Amartha/go-megatron/internal/config"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/acuan"
	lsmPkg "bitbucket.org/Amartha/go-megatron/internal/pkg/lsm"

	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/shopspring/decimal"
)

const (
	LSMRulename     = "LSM Rule"
	LSMRuleFilename = "lsm.grl"
)

type lsm struct {
	rule
	acuanClient acuan.AcuanClient
}

type acuanLSMTransaction struct {
	publisher
	acuan.PublishOrderRequest

	RoundPlaces int32

	AdminFee       decimal.Decimal
	AdminFeeAmount decimal.Decimal
	Percentage     decimal.Decimal
	Tenor          decimal.Decimal

	PPN decimal.Decimal
	A   decimal.Decimal // For calculating PPN (11%)
	B   decimal.Decimal // For calculating PPN (1.11)

	AdminPartnerAccount string
	PPNPartnerAccount   string

	Fee *lsmPkg.Fee

	IsReadyToPublish bool
}

func (t *acuanLSMTransaction) AddTransaction(
	fromAccount,
	toAccount,
	method,
	transactionType,
	currency,
	description string,
	transactionTime time.Time,
	amount decimal.Decimal,
) {
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
		Currency: currency,
	})
}

func (t *acuanLSMTransaction) Publish() {
	t.errPublish = t.publisher.acuanClient.PublishTransaction(t.publisher.ctx, t.PublishOrderRequest)
}

func NewLsm(
	cfg *config.Configuration,
	acuanClient acuan.AcuanClient,
) (Rule[lsmPkg.Event], error) {

	r, err := newRule(cfg, LSMRulename, LSMRuleFilename)
	l := &lsm{
		rule:        r,
		acuanClient: acuanClient,
	}
	return l, err
}

func (r lsm) Execute(ctx context.Context, data lsmPkg.Event) error {
	tx := &acuanLSMTransaction{
		publisher: publisher{
			acuanClient: r.acuanClient,
			ctx:         ctx,
		},
		AdminPartnerAccount: "0",
		PPNPartnerAccount:   "0",
		A:                   decimal.NewFromFloat(0.11),
		B:                   decimal.NewFromFloat(1.11),
		Fee:                 data.Loan.GetAdminFee(),
		Percentage:          decimal.NewFromInt(100),
		Tenor:               decimal.NewFromFloat(data.Loan.Tenor.GetByMonth()).RoundUp(0),
		RoundPlaces:         0,
	}

	dctx := ast.NewDataContext()

	if err := dctx.Add("Loan", &data.Loan); err != nil {
		return err
	}

	if err := dctx.Add("Transaction", tx); err != nil {
		return err
	}

	if err := r.executeEngine(ctx, dctx); err != nil {
		xlog.Info(ctx, "failed to execute with context", xlog.Err(err))
		return err
	}

	if tx.errPublish != nil {
		xlog.Info(ctx, "failed to publish to acuan topic", xlog.Err(tx.errPublish))
		return tx.errPublish
	}

	return nil
}
