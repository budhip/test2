package retry

import (
	"context"
	"errors"

	"bitbucket.org/Amartha/go-megatron/internal/config"
	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/cenkalti/backoff/v4"
)

const DefaultMaxRetries uint64 = 3

type Retryer interface {
	Retry(ctx context.Context, operation, dlqCallback func() error) error
	StopRetryWithErr(err error) error
}

type exponentialBackoff struct {
	ebCfg *config.ExponentialBackOffConfig
}

// NewExponentialBackOff initializes the Retryer with sane defaults.
func NewExponentialBackOff(ebCfg *config.ExponentialBackOffConfig) Retryer {
	if ebCfg.MaxBackoffTime < 0 {
		ebCfg.MaxBackoffTime = backoff.DefaultMaxElapsedTime
	}
	if ebCfg.BackoffMultiplier <= 0 {
		ebCfg.BackoffMultiplier = backoff.DefaultMultiplier
	}
	if ebCfg.MaxRetries <= 0 {
		ebCfg.MaxRetries = DefaultMaxRetries
	}
	return &exponentialBackoff{ebCfg: ebCfg}
}

func (r *exponentialBackoff) Retry(ctx context.Context, operation, dlqCallback func() error) error {
	eb := backoff.NewExponentialBackOff()
	eb.MaxElapsedTime = r.ebCfg.MaxElapsedTime
	eb.InitialInterval = r.ebCfg.InitialInterval
	eb.Multiplier = r.ebCfg.BackoffMultiplier
	eb.RandomizationFactor = 0.0
	// eb.MaxInterval = r.ebCfg.MaxBackoffTime

	var attempt uint64
	wrappedOp := func() error {
		attempt++
		err := operation()
		if err != nil {
			// Context check — skip retries if canceled/deadline exceeded
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				xlog.Warnf(ctx, "[Retry] Context ended on attempt %d: %v", attempt, err)
				return backoff.Permanent(err)
			}
			xlog.Warnf(ctx, "[Retry] Attempt %d failed: %v (will retry)", attempt, err)
		} else {
			xlog.Infof(ctx, "[Retry] Attempt %d succeeded", attempt)
		}
		return err
	}

	err := backoff.Retry(
		wrappedOp,
		backoff.WithContext(backoff.WithMaxRetries(eb, r.ebCfg.MaxRetries), ctx),
	)
	if err != nil {
		xlog.Errorf(ctx, "[Retry] Max retries/time reached (%d attempts). Final error: %v", attempt, err)
		if dlqCallback != nil {
			dlqErr := dlqCallback()
			if dlqErr != nil {
				xlog.Errorf(ctx, "[Retry] DLQ callback failed: %v", dlqErr)
				return dlqErr
			}
			xlog.Infof(ctx, "[Retry] DLQ callback executed successfully.")
		}
		return err
	}

	return nil
}

// StopRetryWithErr stops retrying and returns the provided error.
func (r *exponentialBackoff) StopRetryWithErr(err error) error {
	return backoff.Permanent(err)
}
