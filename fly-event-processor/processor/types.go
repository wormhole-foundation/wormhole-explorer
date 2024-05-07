package processor

import (
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"golang.org/x/net/context"
)

type Params struct {
	TrackID string
	VaaID   string
	ChainID sdk.ChainID
}

// ProcessorFunc is a function to process vaa message.
type ProcessorFunc func(context.Context, *Params) error
