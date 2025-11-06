package graceful

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ProcessStarter func() error

type ProcessStopper func(ctx context.Context) error

func StartProcessAtBackground(ps ...ProcessStarter) {
	for _, p := range ps {
		if p != nil {
			go func(_p func() error) {
				_ = _p()
			}(p)
		}
	}
}

func StopProcessAtBackground(duration time.Duration, ps ...ProcessStopper) {
	sigusr1 := make(chan os.Signal, 1)
	signal.Notify(sigusr1, syscall.SIGUSR1)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	var close = func() {
		for _, p := range ps {
			if p == nil {
				continue
			}
			ctx, stop := context.WithTimeout(context.Background(), duration)
			defer stop()
			_ = p(ctx)
		}
	}
	select {
	case <-sigterm:
		close()
	case <-sigusr1:
		close()
	}
}
