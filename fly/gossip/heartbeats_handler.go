package gossip

import (
	"context"

	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"

	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/health"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/fly/storage"
	"go.uber.org/zap"
)

type heartbeatsHandler struct {
	heartbeatsC chan *gossipv1.Heartbeat
	repository  *storage.Repository
	guardian    *health.GuardianCheck
	metrics     metrics.Metrics
	logger      *zap.Logger
}

func NewHeartbeatsHandler(
	heartbeatsC chan *gossipv1.Heartbeat,
	repository *storage.Repository,
	guardian *health.GuardianCheck,
	metrics metrics.Metrics,
	logger *zap.Logger,
) *heartbeatsHandler {
	return &heartbeatsHandler{
		heartbeatsC: heartbeatsC,
		repository:  repository,
		guardian:    guardian,
		metrics:     metrics,
		logger:      logger,
	}
}

func (h *heartbeatsHandler) Start(ctx context.Context) {
	// Log heartbeats
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case hb := <-h.heartbeatsC:
				h.guardian.Ping(ctx)
				h.metrics.IncHeartbeatFromGossipNetwork(hb.NodeName)
				err := h.repository.UpsertHeartbeat(hb)
				if err != nil {
					h.logger.Error("Error inserting heartbeat", zap.Error(err))
				} else {
					h.metrics.IncHeartbeatInserted(hb.NodeName)
				}
			}
		}
	}()
}
