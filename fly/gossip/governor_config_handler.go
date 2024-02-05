package gossip

import (
	"context"

	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/health"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/fly/storage"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type governorConfigHandler struct {
	govConfigC chan *gossipv1.SignedChainGovernorConfig
	repository *storage.Repository
	guardian   *health.GuardianCheck
	metrics    metrics.Metrics
	logger     *zap.Logger
}

func NewGovernorConfigHandler(
	govConfigC chan *gossipv1.SignedChainGovernorConfig,
	repository *storage.Repository,
	guardian *health.GuardianCheck,
	metrics metrics.Metrics,
	logger *zap.Logger,
) *governorConfigHandler {
	return &governorConfigHandler{
		govConfigC: govConfigC,
		repository: repository,
		guardian:   guardian,
		metrics:    metrics,
		logger:     logger,
	}
}

func (h *governorConfigHandler) Start(ctx context.Context) {
	// Log governor config
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case govConfig := <-h.govConfigC:
				h.guardian.Ping(ctx)
				nodeName, err := h.getGovernorConfigNodeName(govConfig)
				if err != nil {
					h.logger.Error("Error getting gov config node name", zap.Error(err))
					continue
				}
				h.metrics.IncGovernorConfigFromGossipNetwork(nodeName)

				err = h.repository.UpsertGovernorConfig(govConfig)
				if err != nil {
					h.logger.Error("Error inserting gov config", zap.Error(err))
				} else {
					h.metrics.IncGovernorConfigInserted(nodeName)
				}
			}
		}
	}()
}

// getGovernorConfigNodeName get node name from governor config.
func (h *governorConfigHandler) getGovernorConfigNodeName(govConfig *gossipv1.SignedChainGovernorConfig) (string, error) {
	var gCfg gossipv1.ChainGovernorConfig
	err := proto.Unmarshal(govConfig.Config, &gCfg)
	if err != nil {
		return "", err
	}
	return gCfg.NodeName, nil
}
