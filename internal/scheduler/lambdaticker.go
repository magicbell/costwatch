package scheduler

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

type LambdaTicker struct {
	job func(context.Context) error
}

func NewLambdaTicker(job func(context.Context) error) *LambdaTicker {
	return &LambdaTicker{
		job: job,
	}
}

func (l *LambdaTicker) EventBridgeHandler(ctx context.Context, e events.EventBridgeEvent) error {
	if err := l.job(ctx); err != nil {
		return fmt.Errorf("failed to run scheduled job: %w", err)
	}

	return nil
}
