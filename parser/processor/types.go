package processor

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/parser/parser"
)

// ProcessorFunc is a function to process vaa message.
type ProcessorFunc func(context.Context, []byte) (*parser.ParsedVaaUpdate, error)
