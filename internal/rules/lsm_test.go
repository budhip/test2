package rules

import (
	"context"
	"testing"
	"time"

	"bitbucket.org/Amartha/go-megatron/internal/config"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/acuan"
	lsmPkg "bitbucket.org/Amartha/go-megatron/internal/pkg/lsm"
	"github.com/hyperjumptech/grule-rule-engine/pkg"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewLsm(t *testing.T) {
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
			name: "success new lsm rule engine",
			args: args{
				cfg:         th.defaultConfig,
				acuanClient: th.acuanClient,
			},
			doMock: func() {
				th.ruleLoader.EXPECT().
					LoadRule("lsm.grl", "test", Version).
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
					LoadRule("lsm.grl", "test", Version).
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

			_, err := NewLsm(tt.args.cfg, tt.args.acuanClient)
			assert.Equal(t, tt.wantErr, err != nil, err)
		})
	}
}

func TestLsm_Execute(t *testing.T) {
	th := newTestHelper(t)

	RuleLoaderVariable = th.ruleLoader
	defer func() {
		RuleLoaderVariable = &FileRuleLoader{}
	}()

	tests := []struct {
		name     string
		args     lsmPkg.Event
		mockFunc func()
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			args: lsmPkg.Event{
				Loan: lsmPkg.Loan{
					ID:    "1",
					State: "ACTIVE",
					Principal: lsmPkg.Amount{
						Amount: decimal.New(1000000, 10),
					},
				},
			},
			mockFunc: func() {
				mockedRule := []byte(`rule TestRule "test rule" salience 100 {
					when
						true
					then
						Transaction.RefNumber = "12345";
						Transaction.Publish();
						Retract("TestRule");
					}`)

				th.ruleLoader.EXPECT().
					LoadRule("lsm.grl", "test", Version).
					Return(pkg.NewBytesResource(mockedRule), nil)

				th.acuanClient.EXPECT().
					PublishTransaction(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(tt, err)
			},
		},
		{
			name: "error",
			args: lsmPkg.Event{
				Loan: lsmPkg.Loan{
					ID:    "1",
					State: "ACTIVE",
					Principal: lsmPkg.Amount{
						Amount: decimal.New(1000000, 10),
					},
				},
			},
			mockFunc: func() {
				mockedRule := []byte(`rule TestRule "test rule" salience 100 {
					when
						true
					then
						Transaction.RefNumber = "12345";
						Transaction.Publish();
						Retract("TestRule");
					}`)

				th.ruleLoader.EXPECT().
					LoadRule("lsm.grl", "test", Version).
					Return(pkg.NewBytesResource(mockedRule), nil)

				th.acuanClient.EXPECT().
					PublishTransaction(gomock.Any(), gomock.Any()).
					Return(assert.AnError)
			},
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.EqualError(t, err, assert.AnError.Error())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockFunc != nil {
				tt.mockFunc()
			}

			r, err := NewLsm(th.defaultConfig, th.acuanClient)
			assert.NoError(t, err)

			err = r.Execute(context.Background(), tt.args)
			tt.wantErr(t, err)
		})
	}
}

func TestAddTransaction(t *testing.T) {
	trx := &acuanLSMTransaction{}

	trxTime, _ := time.Parse("2006-01-02", "2023-12-12")
	amount, _ := decimal.NewFromString("1000000")

	trx.AddTransaction("000", "123", "", "DSBAB", "IDR", "Testing", trxTime, amount)

	for _, tx := range trx.Transactions {
		assert.Equal(t, tx.Amount, amount)
		assert.Equal(t, tx.Currency, "IDR")
		assert.Equal(t, tx.Description, "Testing")
		assert.Equal(t, tx.FromAccount, "000")
		assert.Equal(t, tx.ToAccount, "123")
		assert.Equal(t, tx.TransactionType, "DSBAB")
	}
}
