package governor

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/domain"
)

type Params struct {
	TrackID         string
	NodeGovernorVaa *domain.NodeGovernorVaa
}

// ProcessorFunc is a function to process a governor message.
type ProcessorFunc func(context.Context, *Params) error
