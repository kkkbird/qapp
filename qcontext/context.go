package qcontext

import (
	"context"
	"errors"
	"time"
)

var (
	ErrDelayedContextTimeout = errors.New("context timeout")
)

// delay 0 meas block forever
func WithDelayedContext(ctx context.Context, delay time.Duration) (context.Context, context.CancelFunc) {
	childCtx, childCancelCause := context.WithCancelCause(context.WithoutCancel(ctx))

	if delay == 0 {
		return childCtx, func() {
			childCancelCause(nil)
		}
	}

	stop := context.AfterFunc(ctx, func() {
		timer := time.NewTimer(delay)
		defer timer.Stop()

		select {
		case <-childCtx.Done():
			return
		case <-timer.C:
			childCancelCause(ErrDelayedContextTimeout)
		}
	})
	return childCtx, func() {
		stop()
		childCancelCause(nil)
	}
}
