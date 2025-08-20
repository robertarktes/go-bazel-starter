package retry

import (
	"context"
	"math"
	"math/rand"
	"time"
)

type Option func(*options)

type options struct {
	maxAttempts int
	backoff     func(attempt int) time.Duration
}

func WithMaxAttempts(n int) Option {
	return func(o *options) { o.maxAttempts = n }
}

func WithBackoff(f func(attempt int) time.Duration) Option {
	return func(o *options) { o.backoff = f }
}

func WithConstantBackoff(d time.Duration) func(attempt int) time.Duration {
	return func(_ int) time.Duration { return d }
}

func WithExponentialBackoff(base time.Duration, factor float64) func(attempt int) time.Duration {
	return func(attempt int) time.Duration {
		return time.Duration(math.Pow(factor, float64(attempt-1))) * base
	}
}

func WithJitterBackoff(base time.Duration, maxJitter time.Duration) func(attempt int) time.Duration {
	return func(_ int) time.Duration {
		return base + time.Duration(rand.Int63n(int64(maxJitter)))
	}
}

func Retry(ctx context.Context, op func() error, opts ...Option) error {
	o := &options{
		maxAttempts: 3,
		backoff:     WithConstantBackoff(1 * time.Second),
	}
	for _, opt := range opts {
		opt(o)
	}

	var err error
	for attempt := 1; attempt <= o.maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err = op()
		if err == nil {
			return nil
		}

		if attempt < o.maxAttempts {
			time.Sleep(o.backoff(attempt))
		}
	}
	return err
}
