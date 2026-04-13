package qcontext

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWithDelayedContext_Comprehensive(t *testing.T) {
	// Tolerance to account for OS scheduler jitter
	const slack = 15 * time.Millisecond

	tests := []struct {
		name          string
		delay         time.Duration
		action        func(cancelParent, cancelChild context.CancelFunc)
		waitDone      bool          // Whether we expect childCtx.Done() to close
		expectedCause error         // What context.Cause(childCtx) should return
		minElapsed    time.Duration // Minimum time expected to pass before Done()
	}{
		{
			name:  "Graceful delay after parent cancellation",
			delay: 100 * time.Millisecond,
			action: func(p, c context.CancelFunc) {
				p() // Cancel parent
			},
			waitDone:      true,
			expectedCause: ErrDelayedContextTimeout,
			minElapsed:    100 * time.Millisecond,
		},
		{
			name:  "Immediate exit on manual child cancellation",
			delay: 1 * time.Second,
			action: func(p, c context.CancelFunc) {
				c() // Cancel child directly
			},
			waitDone:      true,
			expectedCause: context.Canceled, // Manual cancel results in context.Canceled
			minElapsed:    0,
		},
		{
			name:  "Child stays alive if parent is not canceled",
			delay: 100 * time.Millisecond,
			action: func(p, c context.CancelFunc) {
				// Do nothing
			},
			waitDone:      false,
			expectedCause: nil,
		},
		{
			name:  "Parent already canceled at creation",
			delay: 50 * time.Millisecond,
			action: func(p, c context.CancelFunc) {
				p() // Parent is already dead when logic starts
			},
			waitDone:      true,
			expectedCause: ErrDelayedContextTimeout,
			minElapsed:    50 * time.Millisecond,
		},
		{
			name:  "Manual cancel wins during delay period",
			delay: 200 * time.Millisecond,
			action: func(p, c context.CancelFunc) {
				p()                               // Trigger delay
				time.Sleep(50 * time.Millisecond) // Wait for delay to start
				c()                               // Interrupt the delay manually
			},
			waitDone:      true,
			expectedCause: context.Canceled,
			minElapsed:    50 * time.Millisecond, // Should exit at 50ms, not 200ms
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parentCtx, parentCancel := context.WithCancel(context.Background())
			defer parentCancel()

			start := time.Now()
			childCtx, childCancel := WithDelayedContext(parentCtx, tt.delay)
			defer childCancel()

			// Execute the specific test scenario
			tt.action(parentCancel, childCancel)

			if tt.waitDone {
				select {
				case <-childCtx.Done():
					elapsed := time.Since(start)

					// 1. Verify the Cause of cancellation
					gotCause := context.Cause(childCtx)
					if !errors.Is(gotCause, tt.expectedCause) {
						t.Errorf("Cause mismatch: got %v, want %v", gotCause, tt.expectedCause)
					}

					// 2. Verify the timing
					if elapsed < tt.minElapsed-slack {
						t.Errorf("Exited too early: elapsed %v, expected min %v", elapsed, tt.minElapsed)
					}

					// If it was a delay, it shouldn't take MUCH longer than the delay either
					if tt.expectedCause == ErrDelayedContextTimeout && elapsed > tt.delay+slack+50*time.Millisecond {
						t.Errorf("Exited too late: elapsed %v, delay %v", elapsed, tt.delay)
					}

				case <-time.After(tt.delay + 200*time.Millisecond):
					t.Fatalf("Timeout: child context failed to close")
				}
			} else {
				// Ensure it doesn't close prematurely
				select {
				case <-childCtx.Done():
					t.Errorf("Child context closed unexpectedly: %v", context.Cause(childCtx))
				case <-time.After(50 * time.Millisecond):
					// Passed: it stayed open
				}
			}
		})
	}
}
