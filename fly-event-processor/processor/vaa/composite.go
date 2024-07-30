package vaa

import "golang.org/x/net/context"

// CompositeProcessor is a composite processor.
type CompositeProcessor struct {
	processors []ProcessorFunc
}

// NewCompositeProcessor creates a new composite processor.
func NewCompositeProcessor(processors ...ProcessorFunc) *CompositeProcessor {
	return &CompositeProcessor{processors: processors}
}

// Process processes a governor event.
func (c *CompositeProcessor) Process(
	ctx context.Context,
	params *Params) error {
	for _, processor := range c.processors {
		if err := processor(ctx, params); err != nil {
			return err
		}
	}
	return nil
}
