package port

import "context"

// Publisher defines the interface for pushing a payload to a message broker.
type Publisher interface {
	Publish(ctx context.Context, subject string, payload []byte) error
	Close() error
}
