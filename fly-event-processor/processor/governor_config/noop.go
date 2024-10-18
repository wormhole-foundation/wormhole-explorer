package governor_config

import (
	"context"
)

// NoopGovConfigProcessor is a no-op implementation of the Governor Config Processor.
type NoopGovConfigProcessor struct{}

// NewNoopProcessor creates a new NoopGovConfigProcessor.
func NewNoopProcessor() *NoopGovConfigProcessor {
	return &NoopGovConfigProcessor{}
}

// Process is a no-op implementation of the Process method.
func (p *NoopGovConfigProcessor) Process(
	ctx context.Context,
	params *Params) error {
	return nil
}
