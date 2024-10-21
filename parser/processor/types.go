package processor

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/parser/parser"
)

type Params struct {
	Source  string
	TrackID string
	Vaa     []byte
}

// ProcessorFunc is a function to process vaa message.
type ProcessorFunc func(context.Context, *Params) (*parser.ParsedVaaUpdate, error)
