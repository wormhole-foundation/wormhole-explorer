package governor_config

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/queue"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/storage"
	"go.uber.org/zap"
)

// Processor is a governor processor.
type Processor struct {
	repository *storage.PostgresRepository
	logger     *zap.Logger
	metrics    metrics.Metrics
}

// NewProcessor creates a new governor processor.
func NewProcessor(
	repository *storage.PostgresRepository,
	logger *zap.Logger,
	metrics metrics.Metrics,
) *Processor {

	return &Processor{
		repository: repository,
		logger:     logger,
		metrics:    metrics,
	}
}

// Process processes a governor config event.
func (p *Processor) Process(
	ctx context.Context,
	params *Params) error {

	logger := p.logger.With(
		zap.String("trackId", params.TrackID),
	)

	// 1. Get current governor config for the node.
	currentGovernorConfig, err := p.repository.GetGovernorConfig(ctx, params.GovernorConfig.NodeAddress)
	if err != nil {
		logger.Error("failed to get governor config", zap.Error(err))
		return err
	}

	// 2. Check if governor config has changed.
	if !(checkGovernorConfigChanges(currentGovernorConfig, params.GovernorConfig.Chains)) {
		logger.Debug("governor config has not changed",
			zap.String("nodeAddress", params.GovernorConfig.NodeAddress))
		return nil
	}

	// 3 convert governor config to storage model
	var newGovConfigChains []storage.GovernorConfigChain
	for _, chain := range params.GovernorConfig.Chains {
		newGovConfigChains = append(newGovConfigChains, storage.GovernorConfigChain{
			ChainID:            chain.ChainId,
			NotionalLimit:      chain.NotionalLimit,
			BigTransactionSize: chain.BigTransactionSize,
		})
	}

	// 4. Update governor config chains
	err = p.repository.UpdateGovernorConfigChains(ctx, params.GovernorConfig.NodeAddress, newGovConfigChains)
	if err != nil {
		logger.Error("failed to update governor config chains", zap.Error(err))
		return err
	}

	return nil
}

// checkGovernorConfigChanges checks if governor config has changed.
func checkGovernorConfigChanges(
	currentGovernorConfig []storage.GovernorConfigChain,
	newGovernorConfig []*queue.ChainConfig) bool {

	// check if the length of governor config has changed
	if len(currentGovernorConfig) != len(newGovernorConfig) {
		return true
	}

	// convert current governor config to map
	m := make(map[uint16]storage.GovernorConfigChain)
	for _, chain := range currentGovernorConfig {
		m[chain.ChainID] = chain
	}

	// check if new governor config has changed
	var changeGovConfig bool
	for _, n := range newGovernorConfig {
		c, ok := m[n.ChainId]
		if !ok {
			changeGovConfig = true
			break
		}
		if c.BigTransactionSize != n.BigTransactionSize && c.NotionalLimit != n.NotionalLimit {
			changeGovConfig = true
			break
		}
	}
	return changeGovConfig
}
