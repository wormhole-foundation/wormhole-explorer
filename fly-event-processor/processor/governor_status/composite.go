package governor_status

import (
	"context"
)

// Composite is a composite processor.
type CompositeProcessor struct {
	processors []ProcessorFunc
}

// NewComposite creates a new composite processor.
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
