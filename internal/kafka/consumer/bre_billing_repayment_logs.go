package consumer

import (
	"context"
	"fmt"

	"bitbucket.org/Amartha/go-megatron/internal/config"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/accounting"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/acuan"
	brePkg "bitbucket.org/Amartha/go-megatron/internal/pkg/bre"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/graceful"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/kafka"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/metrics"
	"bitbucket.org/Amartha/go-megatron/internal/rules"

	"bitbucket.org/Amartha/go-payment-lib/messaging"
	"bitbucket.org/Amartha/go-payment-lib/messaging/codec"
	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/newrelic/go-agent/v3/newrelic"
)

type breBillingRepaymentLogs struct {
	ctx     context.Context
	sub     messaging.Subscriber
	cfg     *config.Configuration
	breRule rules.Rule[brePkg.Event]
}

func NewBREBillingRepaymentLogs(
	ctx context.Context,
	cfg *config.Configuration,
	nr *newrelic.Application,
	mtc metrics.Metrics,
) (Consumer, []graceful.ProcessStopper, error) {
	var stoppers []graceful.ProcessStopper

	sub, stopper, err := kafka.NewSubscriber(
		cfg.Kafka.Consumers.BREBillingRepaymentLogs.Brokers,
		cfg.Kafka.Consumers.BREBillingRepaymentLogs.ConsumerGroup,
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

	accountingClient := accounting.New(cfg.GoAccounting)

	breRule, err := rules.NewBre(cfg, acuanClient, accountingClient)
	if err != nil {
		return nil, stoppers, err
	}

	c := &breBillingRepaymentLogs{
		ctx:     ctx,
		sub:     sub,
		cfg:     cfg,
		breRule: breRule,
	}

	return c, stoppers, nil
}

func (c *breBillingRepaymentLogs) Start() graceful.ProcessStarter {
	return func() error {
		return c.run()
	}
}

func (c *breBillingRepaymentLogs) run() error {
	err := c.sub.Subscribe(c.ctx,
		messaging.WithTopic(c.cfg.Kafka.Consumers.BREBillingRepaymentLogs.Topic,
			codec.NewJson("v1"),
			c.handler))
	if err != nil {
		return fmt.Errorf("failed subscribing to %s", c.cfg.Kafka.Consumers.BREBillingRepaymentLogs.Topic)
	}

	return nil
}

func (c *breBillingRepaymentLogs) handler(message messaging.Message) messaging.Response {
	var (
		data brePkg.Event
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

	if err = c.breRule.Execute(ctx, data); err != nil {
		return messaging.ReportError(err, nil)
	}

	return messaging.Done(nil)
}
