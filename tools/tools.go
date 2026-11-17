// Package tools provides concrete implementations of agent.Tool.
package tools

import (
	"context"
	"time"
)

const simulatedLatency = 500 * time.Millisecond

func sleepContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
