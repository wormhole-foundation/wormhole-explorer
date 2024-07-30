package governor_config

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/queue"
)

type Params struct {
	TrackID        string
	GovernorConfig queue.GovernorConfig
}

// ProcessorFunc is a function to process a governor message.
type ProcessorFunc func(context.Context, *Params) error
