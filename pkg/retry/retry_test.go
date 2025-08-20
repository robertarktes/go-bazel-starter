package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/robertarktes/go-bazel-starter/pkg/retry"

	"github.com/stretchr/testify/assert"
)

func TestRetry(t *testing.T) {
	tests := []struct {
		name    string
		op      func() error
		opts    []retry.Option
		wantErr bool
	}{
		{
			name:    "success on first try",
			op:      func() error { return nil },
			opts:    []retry.Option{},
			wantErr: false,
		},
		{
			name: "success after retry",
			op: func() func() error {
				count := 0
				return func() error {
					if count < 1 {
						count++
						return errors.New("transient")
					}
					return nil
				}
			}(),
			opts:    []retry.Option{retry.WithMaxAttempts(2), retry.WithBackoff(retry.WithConstantBackoff(10 * time.Millisecond))},
			wantErr: false,
		},
		{
			name:    "context cancellation",
			op:      func() error { return errors.New("fail") },
			opts:    []retry.Option{retry.WithMaxAttempts(3)},
			wantErr: true, // But we'll force cancel
		},
		// Add more
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.name == "context cancellation" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			}
			err := retry.Retry(ctx, tt.op, tt.opts...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
