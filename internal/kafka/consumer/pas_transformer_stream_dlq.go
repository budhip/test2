package consumer

import (
	"context"
	"fmt"

	"bitbucket.org/Amartha/go-megatron/internal/config"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/accounting"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/dddnotification"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/flag"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/graceful"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/kafka"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/metrics"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/pas"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/retry"
	"bitbucket.org/Amartha/go-megatron/internal/rules"
	"bitbucket.org/Amartha/go-payment-lib/messaging"
	"bitbucket.org/Amartha/go-payment-lib/messaging/codec"

	"github.com/newrelic/go-agent/v3/newrelic"
)

type pasTransformerStreamDLQ struct {
	ctx     context.Context
	sub     messaging.Subscriber
	cfg     *config.Configuration
	rule    rules.Rule[pas.OutMessage]
	ebRetry retry.Retryer
}

func NewPASTransformerStreamDLQ(
	ctx context.Context,
	cfg *config.Configuration,
	nr *newrelic.Application,
	mtc metrics.Metrics,
	flag flag.Client,
) (Consumer, []graceful.ProcessStopper, error) {
	var stoppers []graceful.ProcessStopper

	pub, stopperPublisher, err := kafka.NewPublisher(
		cfg.Kafka.Publishers.JournalStream.Brokers,
		cfg.App.Name,
		mtc,
	)
	stoppers = append(stoppers, stopperPublisher)
	if err != nil {
		err = fmt.Errorf("error initiate kafka publisher: %w", err)
		return nil, stoppers, err
	}

	sub, stopperConsumer, err := kafka.NewSubscriber(
		cfg.Kafka.Consumers.TransformerStreamDLQ.Brokers,
		cfg.Kafka.Consumers.TransformerStreamDLQ.ConsumerGroup,
		nr,
		cfg,
		mtc,
	)
	stoppers = append(stoppers, stopperConsumer)
	if err != nil {
		return nil, stoppers, err
	}

	accountingClient := accounting.New(cfg.GoAccounting)
	notificationClient := dddnotification.New(cfg)

	pasRule, err := rules.NewPAS(cfg, pub, accountingClient, flag, notificationClient)
	if err != nil {
		return nil, stoppers, err
	}

	ebRetryer := retry.NewExponentialBackOff(&cfg.ExponentialBackoff)

	c := &pasTransformerStreamDLQ{
		ctx:     ctx,
		sub:     sub,
		rule:    pasRule,
		cfg:     cfg,
		ebRetry: ebRetryer,
	}

	return c, stoppers, nil
}

func (c *pasTransformerStreamDLQ) Start() graceful.ProcessStarter {
	return func() error {
		return c.run()
	}
}

func (c *pasTransformerStreamDLQ) run() error {
	err := c.sub.Subscribe(c.ctx,
		messaging.WithTopic(c.cfg.Kafka.Consumers.TransformerStreamDLQ.Topic,
			codec.NewJson("v1"),
			c.handler))
	if err != nil {
		return fmt.Errorf("failed subscribing to %s", c.cfg.Kafka.Consumers.TransformerStreamDLQ.Topic)
	}

	return nil
}

func (c *pasTransformerStreamDLQ) handler(message messaging.Message) messaging.Response {
	var (
		out pas.OutMessage
		err error
	)
	ctx := message.Context()

	if err = message.Bind(&out); err != nil {
		return messaging.ExpectError(err, nil)
	}

	var operationErr error
	operation := func() error {
		operationErr = c.rule.Execute(ctx, out)
		if operationErr != nil {
			return operationErr
		}
		return nil
	}

	if err = c.ebRetry.Retry(ctx, operation, nil); err != nil {
		return messaging.ReportError(err, out)
	}

	return messaging.Done(out)
}
