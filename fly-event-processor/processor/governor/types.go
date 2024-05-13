package governor

import "context"

type Params struct {
	TrackID string
}

// ProcessorFunc is a function to process a governor message.
type ProcessorFunc func(context.Context, *Params) error
