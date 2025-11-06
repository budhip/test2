package rules

import (
	"context"
	"testing"
	"time"

	"bitbucket.org/Amartha/go-megatron/internal/config"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/acuan"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/acuan/mock"

	mockAccounting "bitbucket.org/Amartha/go-megatron/internal/pkg/accounting/mock"
	mockNotification "bitbucket.org/Amartha/go-megatron/internal/pkg/dddnotification/mock"
	mockKafka "bitbucket.org/Amartha/go-megatron/internal/pkg/kafka/mock"
	papaModel "bitbucket.org/Amartha/go-megatron/internal/pkg/papa"
	mockLoader "bitbucket.org/Amartha/go-megatron/internal/rules/mock"
	paymentLib "bitbucket.org/Amartha/go-payment-lib/payment-api/models"

	"github.com/hyperjumptech/grule-rule-engine/pkg"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type testHelper struct {
	acuanClient        *mock.MockAcuanClient
	accountingClient   *mockAccounting.MockClient
	notificationClient *mockNotification.MockClient
	publisher          *mockKafka.MockPublisher
	ruleLoader         *mockLoader.MockRuleLoader

	defaultRule   []byte
	defaultConfig *config.Configuration
}

func newTestHelper(t *testing.T) testHelper {
	t.Helper()

	mockCtrl := gomock.NewController(t)
	acuanClient := mock.NewMockAcuanClient(mockCtrl)
	accountingClient := mockAccounting.NewMockClient(mockCtrl)
	notificationClient := mockNotification.NewMockClient(mockCtrl)
	publisher := mockKafka.NewMockPublisher(mockCtrl)
	ruleLoader := mockLoader.NewMockRuleLoader(mockCtrl)

	defaultRule := []byte(`
	rule TestRule "test rule" salience 100 {
    when
        true
    then
		Log("Rule Executed");
        Retract("TestRule");
	}`)

	return testHelper{
		acuanClient:        acuanClient,
		accountingClient:   accountingClient,
		notificationClient: notificationClient,
		publisher:          publisher,
		ruleLoader:         ruleLoader,
		defaultRule:        defaultRule,
		defaultConfig: &config.Configuration{
			App: config.AppConfiguration{
				Env: "test",
			},
		},
	}
}

func TestNewPapa(t *testing.T) {
	th := newTestHelper(t)

	RuleLoaderVariable = th.ruleLoader
	defer func() {
		RuleLoaderVariable = &FileRuleLoader{}
	}()

	type args struct {
		cfg         *config.Configuration
		acuanClient acuan.AcuanClient
	}
	tests := []struct {
		name    string
		args    args
		doMock  func()
		wantErr bool
	}{
		{
			name: "success new papa rule engine",
			args: args{
				cfg:         th.defaultConfig,
				acuanClient: th.acuanClient,
			},
			doMock: func() {
				th.ruleLoader.EXPECT().
					LoadRule("papa.grl", "test", Version).
					Return(pkg.NewBytesResource(th.defaultRule), nil)
			},
		},
		{
			name: "error load rule",
			args: args{
				cfg:         th.defaultConfig,
				acuanClient: th.acuanClient,
			},
			doMock: func() {
				th.ruleLoader.EXPECT().
					LoadRule("papa.grl", "test", Version).
					Return(nil, assert.AnError)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock()
			}

			_, err := NewPapa(tt.args.cfg, tt.args.acuanClient, nil)
			assert.Equal(t, tt.wantErr, err != nil, err)
		})
	}
}

func Test_papa_Execute(t *testing.T) {
	th := newTestHelper(t)

	RuleLoaderVariable = th.ruleLoader
	defer func() {
		RuleLoaderVariable = &FileRuleLoader{}
	}()

	defaultSuccessTransactionTime := paymentLib.DateTime(time.Now().Format(time.RFC3339))

	type args struct {
		ctx  context.Context
		data papaModel.TransactionStreamEvent
	}
	tests := []struct {
		name    string
		args    args
		doMock  func()
		wantErr bool
	}{
		{
			name: "success execute rule",
			args: args{
				ctx: context.Background(),
				data: papaModel.TransactionStreamEvent{
					Transaction: paymentLib.Transaction{
						Amount:       "100000",
						SuccessfulAt: &defaultSuccessTransactionTime,
					},
				},
			},
			doMock: func() {
				mockedRule := []byte(`rule TestRule "test rule" salience 100 {
					when
						true
					then
						Transaction.Amount = Payment.Amount;
						Transaction.TransactionTime = Payment.SuccessfulAt;
						Order.Transactions.Append(Transaction.GetValue());
						Order.Publish();
						Retract("TestRule");
					}`)

				th.ruleLoader.EXPECT().
					LoadRule("papa.grl", "test", Version).
					Return(pkg.NewBytesResource(mockedRule), nil)

				th.acuanClient.EXPECT().
					PublishTransaction(gomock.Any(), gomock.Any()).
					Return(nil)
			},
		},
		{
			name: "success - append 2 transaction with metadata",
			args: args{
				ctx: context.Background(),
				data: papaModel.TransactionStreamEvent{
					Transaction: paymentLib.Transaction{
						Amount:       "100000",
						SuccessfulAt: &defaultSuccessTransactionTime,
					},
				},
			},
			doMock: func() {
				mockedRule := []byte(`
					rule SetDefaultAttribute "Set default attributes" salience 100 {
						when
							!Transaction.DefaultAttributeSet
						then
							Transaction.TransactionTime = Payment.SuccessfulAt;
							Transaction.Amount = Payment.Amount;

							Transaction.DefaultAttributeSet = true;
					}

					rule TestAppendFirstData "test append first data" salience 99 {
						when
							true
						then
							Metadata.Description = "This is a first data test";
							Order.Transactions.Append(Transaction.GetValue());
							Changed("Transaction");
							
							Transaction.DefaultAttributeSet = false;
							Retract("TestAppendFirstData");
					}

					rule TestAppendSecondData "test append second data" salience 98 {
						when
							true
						then
							Metadata.Description = "This is a second data test";
							Order.Transactions.Append(Transaction.GetValue());
							
							Order.Publish();
							Retract("TestAppendSecondData");
					}
`)

				th.ruleLoader.EXPECT().
					LoadRule("papa.grl", "test", Version).
					Return(pkg.NewBytesResource(mockedRule), nil)

				th.acuanClient.EXPECT().
					PublishTransaction(gomock.Any(), gomock.Any()).
					Return(nil)
			},
		},
		{
			name: "error execute rule",
			args: args{
				ctx: context.Background(),
				data: papaModel.TransactionStreamEvent{
					Transaction: paymentLib.Transaction{
						Amount:       "100000",
						SuccessfulAt: &defaultSuccessTransactionTime,
					},
				},
			},
			doMock: func() {
				mockedRule := []byte(`rule TestRule "test rule" salience 100 {
					when
						true
					then
						Order.NoFieldMath = Payment.NoFieldMath;
						Order.Publish();
						Retract("TestRule");
					}`)

				th.ruleLoader.EXPECT().
					LoadRule("papa.grl", "test", Version).
					Return(pkg.NewBytesResource(mockedRule), nil)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock()
			}

			r, err := NewPapa(th.defaultConfig, th.acuanClient, nil)
			assert.NoError(t, err)

			err = r.Execute(tt.args.ctx, tt.args.data)
			assert.Equal(t, tt.wantErr, err != nil, err)
		})
	}
}

func Test_acuanPAPATransformed_Publish(t1 *testing.T) {
	th := newTestHelper(t1)

	defaultSuccessTransactionTime := paymentLib.DateTime(time.Now().Format(time.RFC3339))

	RuleLoaderVariable = th.ruleLoader
	defer func() {
		RuleLoaderVariable = &FileRuleLoader{}
	}()

	tests := []struct {
		name           string
		fields         *acuanPAPATransformedOrder
		doMock         func()
		wantErrPublish bool
	}{
		{
			name: "success publish",
			fields: &acuanPAPATransformedOrder{
				Transactions: []acuanPAPATransaction{
					{
						Amount:          "100000",
						TransactionTime: &defaultSuccessTransactionTime,
					},
				},
				publisher: publisher{
					acuanClient: th.acuanClient,
					ctx:         context.TODO(),
				},
			},
			doMock: func() {
				th.acuanClient.EXPECT().
					PublishTransaction(gomock.Any(), gomock.Any()).
					Return(nil)
			},
		},
		{
			name: "error publish",
			fields: &acuanPAPATransformedOrder{
				Transactions: []acuanPAPATransaction{
					{
						Amount:          "100000",
						TransactionTime: &defaultSuccessTransactionTime,
					},
				},
				publisher: publisher{
					acuanClient: th.acuanClient,
					ctx:         context.TODO(),
				},
			},
			doMock: func() {
				th.acuanClient.EXPECT().
					PublishTransaction(gomock.Any(), gomock.Any()).
					Return(assert.AnError)
			},
			wantErrPublish: true,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			if tt.doMock != nil {
				tt.doMock()
			}

			tt.fields.Publish()
			assert.Equal(t1, tt.wantErrPublish, tt.fields.errPublish != nil, tt.fields.errPublish)
		})
	}
}

func Test_acuanPAPATransformed_toPublishOrderRequest(t1 *testing.T) {
	th := newTestHelper(t1)

	ct := time.Now()
	amount := "10000"
	defaultSuccessTransactionTime := paymentLib.DateTime(ct.Format(time.RFC3339))

	RuleLoaderVariable = th.ruleLoader
	defer func() {
		RuleLoaderVariable = &FileRuleLoader{}
	}()

	tests := []struct {
		name    string
		fields  *acuanPAPATransformedOrder
		wantErr bool
	}{
		{
			name: "success transform to publish order request",
			fields: &acuanPAPATransformedOrder{
				Transactions: []acuanPAPATransaction{
					{
						Amount:          paymentLib.Amount(amount),
						TransactionTime: &defaultSuccessTransactionTime,
					},
				},
				publisher: publisher{
					acuanClient: th.acuanClient,
					ctx:         context.TODO(),
				},
			},
		},
		{
			name: "error - amount is not decimal",
			fields: &acuanPAPATransformedOrder{
				Transactions: []acuanPAPATransaction{
					{
						Amount:          "INVALID_AMOUNT",
						TransactionTime: &defaultSuccessTransactionTime,
					},
				},
				publisher: publisher{
					acuanClient: th.acuanClient,
					ctx:         context.TODO(),
				},
			},
			wantErr: true,
		},
		{
			name: "error - transaction time is nil",
			fields: &acuanPAPATransformedOrder{
				Transactions: []acuanPAPATransaction{
					{
						Amount:          paymentLib.Amount(amount),
						TransactionTime: nil,
					},
				},
				publisher: publisher{
					acuanClient: th.acuanClient,
					ctx:         context.TODO(),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			_, err := tt.fields.toPublishOrderRequest()
			assert.Equal(t1, tt.wantErr, err != nil, err)
		})
	}
}
