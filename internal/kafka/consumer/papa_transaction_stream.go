package consumer

import (
	"context"
	"fmt"

	"bitbucket.org/Amartha/go-megatron/internal/config"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/acuan"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/flag"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/graceful"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/kafka"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/metrics"
	papaModel "bitbucket.org/Amartha/go-megatron/internal/pkg/papa"
	"bitbucket.org/Amartha/go-megatron/internal/rules"

	"bitbucket.org/Amartha/go-payment-lib/messaging"
	"bitbucket.org/Amartha/go-payment-lib/messaging/codec"
	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/newrelic/go-agent/v3/newrelic"
)

type papaTransactionStream struct {
	ctx  context.Context
	sub  messaging.Subscriber
	cfg  *config.Configuration
	rule rules.Rule[papaModel.TransactionStreamEvent]
}

func NewPAPATransactionStream(
	ctx context.Context,
	cfg *config.Configuration,
	nr *newrelic.Application,
	mtc metrics.Metrics,
	flag flag.Client,
) (Consumer, []graceful.ProcessStopper, error) {
	var stoppers []graceful.ProcessStopper

	sub, stopper, err := kafka.NewSubscriber(
		cfg.Kafka.Consumers.PAPATransactionStream.Brokers,
		cfg.Kafka.Consumers.PAPATransactionStream.ConsumerGroup,
		nr,
		cfg,
		mtc,
	)
	stoppers = append(stoppers, stopper)
	if err != nil {
		return nil, stoppers, err
	}

	acuanClient, err := acuan.New(cfg)
	if err != nil {
		return nil, stoppers, err
	}

	papaRule, err := rules.NewPapa(cfg, acuanClient, flag)
	if err != nil {
		return nil, stoppers, err
	}

	c := &papaTransactionStream{
		ctx:  ctx,
		sub:  sub,
		rule: papaRule,
		cfg:  cfg,
	}

	return c, stoppers, nil
}

func (c *papaTransactionStream) Start() graceful.ProcessStarter {
	return func() error {
		return c.run()
	}
}

func (c *papaTransactionStream) run() error {
	err := c.sub.Subscribe(c.ctx,
		messaging.WithTopic(c.cfg.Kafka.Consumers.PAPATransactionStream.Topic,
			codec.NewJson("v1"),
			c.handler))
	if err != nil {
		return fmt.Errorf("failed subscribing to %s", c.cfg.Kafka.Consumers.PAPATransactionStream.Topic)
	}

	return nil
}

func (c *papaTransactionStream) handler(message messaging.Message) messaging.Response {
	var (
		data papaModel.TransactionStreamEvent
		err  error
	)
	ctx := message.Context()

	defer func() {
		if err != nil {
			xlog.Error(ctx, logConsumerHandler, xlog.Err(err))
		}
	}()

	if err = message.Bind(&data); err != nil {
		return messaging.ExpectError(err, nil)
	}

	if err = c.rule.Execute(ctx, data); err != nil {
		return messaging.ReportError(err, nil)
	}

	return messaging.Done(nil)
}
