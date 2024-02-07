package gossip

import (
	"context"

	"github.com/certusone/wormhole/node/pkg/common"
	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/health"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/fly/processor"
)

type observationHandler struct {
	obsvC    chan *common.MsgWithTimeStamp[gossipv1.SignedObservation]
	pushFunc processor.ObservationPushFunc
	guardian *health.GuardianCheck
	metrics  metrics.Metrics
}

func NewObservationHandler(
	obsvC chan *common.MsgWithTimeStamp[gossipv1.SignedObservation],
	pushFunc processor.ObservationPushFunc,
	guardian *health.GuardianCheck,
	metrics metrics.Metrics,
) *observationHandler {
	return &observationHandler{
		obsvC:    obsvC,
		pushFunc: pushFunc,
		guardian: guardian,
		metrics:  metrics,
	}
}

func (h *observationHandler) Start(ctx context.Context) {
	// Log observations
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case m := <-h.obsvC:
				o := m.Msg
				h.guardian.Ping(ctx)
				h.metrics.IncObservationTotal()
				h.pushFunc(ctx, o)
			}
		}
	}()
}
