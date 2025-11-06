package rules

import (
	"context"
	"errors"
	"fmt"

	paymentLib "bitbucket.org/Amartha/go-payment-lib/payment-api/models"

	"bitbucket.org/Amartha/go-megatron/internal/config"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/acuan"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/flag"
	papaModel "bitbucket.org/Amartha/go-megatron/internal/pkg/papa"

	"github.com/hyperjumptech/grule-rule-engine/ast"
)

type papa struct {
	rule
	acuanClient acuan.AcuanClient
	flagClient  flag.Client
}

type publisher struct {
	// errPublish is error information when publishing to Acuan
	errPublish  error
	acuanClient acuan.AcuanClient
	ctx         context.Context
}

type metadataPayment struct {
	Description                               string            `json:"description"`
	PaymentTransactionID                      string            `json:"paymentTransactionID"`
	PaymentTransactionReferenceNumber         string            `json:"paymentTransactionReferenceNumber"`
	PaymentTransactionExternalReferenceNumber string            `json:"paymentTransactionExternalReferenceNumber"`
	EntityCode                                paymentLib.Entity `json:"entityCode"`
	CustomerName                              string            `json:"customerName"`
	AccountNumber                             string            `json:"accountNumber"`

	Amount   paymentLib.Amount `json:"amount,omitempty"`
	AdminFee paymentLib.Amount `json:"adminFee,omitempty"`
	Net      paymentLib.Amount `json:"net,omitempty"`
}

// acuanPAPATransaction is struct to be used to transform payment transaction to acuan transaction
type acuanPAPATransaction struct {
	FromAccount     string
	ToAccount       string
	TransactionType string

	// TransactionTime is successful transaction time from payment transaction, we can't use time.Time since grule
	// doesn't support assignment from different data type. This also applies to Amount and Currency
	TransactionTime *paymentLib.DateTime
	Amount          paymentLib.Amount
	Currency        paymentLib.Currency
	Description     string
	Metadata        metadataPayment

	// IsDefaultAttributeSet is flag to check from rule engine if default attribute is set
	DefaultAttributeSet bool
}

type flagClient struct {
	client flag.Client
}

func (f *flagClient) IsEnabled(key string) bool {
	return f.client.IsEnabled(key)
}

// GetValue return non-pointer value of acuanPAPATransaction so, it can be used as argument in rule engine
func (transaction *acuanPAPATransaction) GetValue() acuanPAPATransaction {
	return *transaction
}

// acuanPAPATransformedOrder is struct to be used as data context in rule engine
// it will be used to transform payment transaction to acuan publish order request.
// 1 transaction PAPA may have multiple acuan Transactions
type acuanPAPATransformedOrder struct {
	publisher

	RefNumber    string
	OrderType    string
	Transactions []acuanPAPATransaction

	// IsReadyToPublish is flag to determine if transaction is ready to publish to Acuan
	IsReadyToPublish bool
}

func (order *acuanPAPATransformedOrder) toPublishOrderRequest() (trx acuan.PublishOrderRequest, err error) {
	var transactions []acuan.OrderTransaction
	for _, transaction := range order.Transactions {
		amount, errParse := transaction.Amount.ToDecimal()
		if errParse != nil {
			return trx, fmt.Errorf("error converting amount to decimal: %w", errParse)
		}

		if amount.IsZero() {
			continue
		}

		if transaction.TransactionTime == nil {
			return trx, errors.New("failed to create acuan request data, transaction time is nil")
		}

		transactionTime, errParse := transaction.TransactionTime.RFC3339()
		if errParse != nil {
			return trx, fmt.Errorf("error converting transaction time: %w", errParse)
		}

		transactions = append(transactions, acuan.OrderTransaction{
			FromAccount:     transaction.FromAccount,
			ToAccount:       transaction.ToAccount,
			Amount:          amount,
			TransactionType: transaction.TransactionType,
			TransactionTime: transactionTime,
			Description:     transaction.Description,
			Metadata:        transaction.Metadata,
			Currency:        string(transaction.Currency),
		})
	}

	return acuan.PublishOrderRequest{
		OrderType:    order.OrderType,
		RefNumber:    order.RefNumber,
		Transactions: transactions,
	}, nil
}

func (order *acuanPAPATransformedOrder) Publish() {
	payload, err := order.toPublishOrderRequest()
	if err != nil {
		order.errPublish = err
		return
	}

	order.errPublish = order.publisher.acuanClient.PublishTransaction(order.publisher.ctx, payload)
}

func NewPapa(
	cfg *config.Configuration,
	acuanClient acuan.AcuanClient,
	flagClient flag.Client) (Rule[papaModel.TransactionStreamEvent], error) {

	r, err := newRule(cfg, "PAPA Rule", "papa.grl")
	p := &papa{
		rule:        r,
		acuanClient: acuanClient,
		flagClient:  flagClient,
	}
	return p, err
}

func (r papa) Execute(ctx context.Context, data papaModel.TransactionStreamEvent) error {
	dctx := ast.NewDataContext()

	err := dctx.Add("Payment", &data)
	if err != nil {
		return err
	}

	err = dctx.Add("Flag", &flagClient{
		r.flagClient,
	})
	if err != nil {
		return err
	}

	odr := &acuanPAPATransformedOrder{
		publisher: publisher{
			ctx:         ctx,
			acuanClient: r.acuanClient,
		},
	}
	err = dctx.Add("Order", odr)
	if err != nil {
		return err
	}

	tx := &acuanPAPATransaction{}
	err = dctx.Add("Transaction", tx)
	if err != nil {
		return err
	}

	err = dctx.Add("Metadata", &tx.Metadata)
	if err != nil {
		return err
	}

	err = r.executeEngine(ctx, dctx)
	if err != nil {
		return err
	}

	return odr.errPublish
}
