package consumer

import (
	"context"

	"bitbucket.org/Amartha/go-megatron/internal/config"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/flag"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/graceful"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/metrics"

	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/newrelic/go-agent/v3/newrelic"
)

const (
	logConsumerHandler       = "[CONSUMER-HANDLER]"
	LSMLoanLogs              = "loan_logs"
	PAPATransactionStream    = "transaction_stream"
	BREBillingRepaymentLogs  = "billing_repayment_logs"
	PASAcuanTransactionNotif = "acuan_transaction_notif"
	PASTransformerStreamDLQ  = "transformer_stream_dlq"
)

var (
	ListConsumerName = []string{
		LSMLoanLogs,
		PAPATransactionStream,
		BREBillingRepaymentLogs,
		PASAcuanTransactionNotif,
		PASTransformerStreamDLQ,
	}
)

type Consumer interface {
	Start() graceful.ProcessStarter
}

func New(
	ctx context.Context,
	name string,
	cfg *config.Configuration,
	nr *newrelic.Application,
	mtc metrics.Metrics,
	flag flag.Client,
) (Consumer, []graceful.ProcessStopper, error) {
	switch name {
	case PAPATransactionStream:
		return NewPAPATransactionStream(ctx, cfg, nr, mtc, flag)
	case LSMLoanLogs:
		return NewLSMLoanLogs(ctx, cfg, nr, mtc)
	case PASAcuanTransactionNotif:
		return NewPASAcuanTransactionNotif(ctx, cfg, nr, mtc, flag)
	case PASTransformerStreamDLQ:
		return NewPASTransformerStreamDLQ(ctx, cfg, nr, mtc, flag)
	case BREBillingRepaymentLogs:
		return NewBREBillingRepaymentLogs(ctx, cfg, nr, mtc)
	default:
		xlog.Error(ctx, "invalid consumer instance name")
	}
	return nil, nil, nil
}
