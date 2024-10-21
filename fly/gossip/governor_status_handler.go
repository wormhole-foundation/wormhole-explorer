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

type governorStatusHandler struct {
	govStatusC chan *gossipv1.SignedChainGovernorStatus
	repository storage.Storager
	guardian   *health.GuardianCheck
	metrics    metrics.Metrics
	logger     *zap.Logger
}

func NewGovernorStatusHandler(
	govStatusC chan *gossipv1.SignedChainGovernorStatus,
	repository storage.Storager,
	guardian *health.GuardianCheck,
	metrics metrics.Metrics,
	logger *zap.Logger,
) *governorStatusHandler {
	return &governorStatusHandler{
		govStatusC: govStatusC,
		repository: repository,
		guardian:   guardian,
		metrics:    metrics,
		logger:     logger,
	}
}

func (h *governorStatusHandler) Start(ctx context.Context) {
	// Log govStatus
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case govStatus := <-h.govStatusC:
				h.guardian.Ping(ctx)
				nodeName, err := h.getGovernorStatusNodeName(govStatus)
				if err != nil {
					h.logger.Error("Error getting gov status node name", zap.Error(err))
					continue
				}
				h.metrics.IncGovernorStatusFromGossipNetwork(nodeName)
				err = h.repository.UpsertGovernorStatus(ctx, govStatus)
				if err != nil {
					h.logger.Error("Error inserting gov status", zap.Error(err))
				} else {
					h.metrics.IncGovernorStatusInserted(nodeName)
				}
			}
		}
	}()
}

// getGovernorStatusNodeName get node name from governor status.
func (h *governorStatusHandler) getGovernorStatusNodeName(govStatus *gossipv1.SignedChainGovernorStatus) (string, error) {
	var gStatus gossipv1.ChainGovernorStatus
	err := proto.Unmarshal(govStatus.Status, &gStatus)
	if err != nil {
		return "", err
	}
	return gStatus.NodeName, nil
}
