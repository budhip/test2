package kafka

import (
	"context"
	"fmt"
	"time"

	"bitbucket.org/Amartha/go-megatron/internal/config"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/graceful"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/metrics"

	"bitbucket.org/Amartha/go-payment-lib/messaging"
	"bitbucket.org/Amartha/go-payment-lib/messaging/kafka"
	"bitbucket.org/Amartha/go-payment-lib/messaging/kafka/middleware"
	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/newrelic/go-agent/v3/newrelic"
)

const logField = "[KAFKA-CONSUMER]"

type Subscriber interface {
	messaging.Subscriber
}

func NewSubscriber(
	brokers []string,
	consumerGroup string,
	nr *newrelic.Application,
	cfg *config.Configuration,
	mtc metrics.Metrics,
) (Subscriber, graceful.ProcessStopper, error) {
	stopper := func(context.Context) error { return nil }

	defaultOpts := []kafka.SubscriberOption{
		kafka.WithSubscriberKafkaVersion("3.3.1"),
		kafka.WithMiddleware(
			middleware.Context,
			middlewareMessageClaim,
		),
		kafka.WithSubscriberPreStartCallback(preStart),
		kafka.WithSubscriberEndCallback(end),
		kafka.WithSubscriberErrorCallback(errcb),
		kafka.WithSubscriberSessionTimeout(5 * time.Minute),
	}
	if mtc != nil {
		defaultOpts = append(
			defaultOpts,
			kafka.WithSubscriberGenericPromMetrics(
				mtc.PrometheusRegisterer(),
				fmt.Sprintf("%s_consumer_metric", cfg.App.Name),
				consumerGroup,
				1*time.Second,
			),
		)
	}

	sub, err := kafka.NewSubscriber(
		brokers,
		consumerGroup,
		defaultOpts...,
	)
	if err != nil {
		return nil, stopper, err
	}

	stopper = func(ctx context.Context) error {
		return sub.CloseSubscriber()
	}

	return sub, stopper, nil
}

func GetMessageClaim(message messaging.Message) kafka.MessageClaim {
	v, ok := message.GetMessageClaim().(kafka.MessageClaim)
	if !ok {
		return nil
	}
	return v
}

func middlewareMessageClaim(next messaging.SubscriptionHandler) messaging.SubscriptionHandler {
	return func(message messaging.Message) messaging.Response {
		v := GetMessageClaim(message)
		if v != nil {
			logField := []xlog.Field{
				xlog.Time("timestamp", v.Timestamp),
				xlog.Any("block-timestamp", v.BlockTimestamp),
				xlog.String("topic", v.Topic),
				xlog.Int32("partition", v.Partition),
				xlog.Int64("offset", v.Offset),
				xlog.Any("header", v.Headers),
				xlog.String("key", string(v.Key)),
				xlog.String("message-claimed", string(v.Value)),
			}
			xlog.Info(message.Context(), "[MESSAGE-CLAIM]", logField...)
		}

		msg := next(message)
		if msg.IsError() {
			xlog.Error(message.Context(), logField, xlog.Any("error", msg.Error()))
		}

		return msg
	}
}

func preStart(ctx context.Context, claims map[string][]int32) error {
	xlog.Info(ctx, logField, xlog.String("status", "start"), xlog.Any("claims", claims))
	return nil
}

func end(ctx context.Context, claims map[string][]int32) error {
	xlog.Info(ctx, logField, xlog.String("status", "end"), xlog.Any("claims", claims))
	return nil
}

func errcb(err error) {
	xlog.Warn(context.Background(), logField, xlog.String("status", "error consumer group"), xlog.Err(err))
}
