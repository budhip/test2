package rules

import (
	"context"
	"testing"

	"bitbucket.org/Amartha/go-megatron/internal/config"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/kafka"
	paspkg "bitbucket.org/Amartha/go-megatron/internal/pkg/pas"

	"github.com/hyperjumptech/grule-rule-engine/pkg"
	"github.com/stretchr/testify/assert"
)

func TestNewPAS(t *testing.T) {
	th := newTestHelper(t)

	RuleLoaderVariable = th.ruleLoader
	defer func() {
		RuleLoaderVariable = &FileRuleLoader{}
	}()

	type args struct {
		cfg       *config.Configuration
		publisher kafka.Publisher
	}
	tests := []struct {
		name    string
		args    args
		doMock  func()
		wantErr bool
	}{
		{
			name: "success new pas rule engine",
			args: args{
				cfg:       th.defaultConfig,
				publisher: th.publisher,
			},
			doMock: func() {
				th.ruleLoader.EXPECT().
					LoadRule("pas.grl", "test", Version).
					Return(pkg.NewBytesResource(th.defaultRule), nil)
			},
		},
		{
			name: "error load rule",
			args: args{
				cfg:       th.defaultConfig,
				publisher: th.publisher,
			},
			doMock: func() {
				th.ruleLoader.EXPECT().
					LoadRule("pas.grl", "test", Version).
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
			_, err := NewPAS(tt.args.cfg, tt.args.publisher, th.accountingClient, nil, th.notificationClient)
			assert.Equal(t, tt.wantErr, err != nil, err)
		})
	}
}

func Test_PAS_Execute(t *testing.T) {
	th := newTestHelper(t)

	RuleLoaderVariable = th.ruleLoader
	defer func() {
		RuleLoaderVariable = &FileRuleLoader{}
	}()

	type args struct {
		ctx  context.Context
		data paspkg.OutMessage
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
				data: paspkg.OutMessage{
					Status: "SUCCESS",
				},
			},
			doMock: func() {
				mockedRule := []byte(`
				rule TransformAcuanTransactionDisbursementPLBroilerX "When it's disbursement partnership loan broilerX, transform to transaction PAS" salience 10 {
					when
						Acuan.AcuanStatus == "SUCCESS"
					then
						Journal.TransactionID = Acuan.Id;
						Journal.OrderType = Acuan.OrderType;
						Journal.TransactionDate = Acuan.TransactionTime;
						Journal.Currency = Acuan.Currency;
						Journal.Metadata = Acuan.Meta;
				
						JournalDebit.TransactionType = Acuan.TransactionType;
						JournalDebit.Account = "121001000000003";
						JournalDebit.Narrative = "Testing";
						JournalDebit.Amount = Acuan.Amount;
						JournalDebit.IsDebit = true;
						Journal.Transactions[0] = JournalDebit;
				
						JournalCredit.TransactionType = Acuan.TransactionType;
						JournalCredit.Account = Acuan.DestinationAccountId;
						JournalCredit.Narrative = "Testing";
						JournalCredit.Amount = Acuan.Amount;
						JournalCredit.IsDebit = false;
						Journal.Transactions[1] = JournalCredit;
						Transaction.IsReadyToPublish = true;
						Transaction.Publish();
						Retract("TransformAcuanTransactionDisbursementPLBroilerX");
				}`)
				th.ruleLoader.EXPECT().
					LoadRule("pas.grl", "test", Version).
					Return(pkg.NewBytesResource(mockedRule), nil)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.doMock != nil {
				tt.doMock()
			}

			r, err := NewPAS(th.defaultConfig, th.publisher, th.accountingClient, nil, th.notificationClient)
			assert.NoError(t, err)

			err = r.Execute(tt.args.ctx, tt.args.data)
			assert.Equal(t, tt.wantErr, err != nil, err)
		})
	}
}
