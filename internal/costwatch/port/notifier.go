package port

import "context"

// Notifier delivers alert messages.
type Notifier interface {
	Send(ctx context.Context, text string) error
}
