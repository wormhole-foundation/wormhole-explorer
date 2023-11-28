package producer

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/common/events"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// PushFunc is a function to push VAAEvent.
type PushFunc func(context.Context, *Notification) error

type Notification struct {
	ID           string
	Event        *events.NotificationEvent
	EmitterChain sdk.ChainID
}

// NewComposite returns a PushFunc that calls all the given producers.
func NewComposite(producers ...PushFunc) PushFunc {
	return func(ctx context.Context, event *Notification) error {
		for _, producer := range producers {
			if err := producer(ctx, event); err != nil {
				return err
			}
		}
		return nil
	}
}
