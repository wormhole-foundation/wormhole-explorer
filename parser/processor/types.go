package processor

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/parser/parser"
)

// ProcessorFunc is a function to process ParsedVaaUpdate
type ProcessorFunc func(context.Context, *parser.ParsedVaaUpdate) error
