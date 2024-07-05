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
	obsvC            chan *common.MsgWithTimeStamp[gossipv1.SignedObservation]
	batchObsvC       chan *common.MsgWithTimeStamp[gossipv1.SignedObservationBatch]
	pushObsFunc      processor.ObservationPushFunc
	pushBatchObsFunc processor.BatchObservationPushFunc
	guardian         *health.GuardianCheck
	metrics          metrics.Metrics
}

func NewObservationHandler(obsvC chan *common.MsgWithTimeStamp[gossipv1.SignedObservation], batchObsvC chan *common.MsgWithTimeStamp[gossipv1.SignedObservationBatch], pushFunc processor.ObservationPushFunc, pushBatchObsFunc func(_ context.Context, batchMsg *gossipv1.SignedObservationBatch) error, guardian *health.GuardianCheck, metrics metrics.Metrics) *observationHandler {
	return &observationHandler{
		obsvC:            obsvC,
		pushObsFunc:      pushFunc,
		guardian:         guardian,
		metrics:          metrics,
		batchObsvC:       batchObsvC,
		pushBatchObsFunc: pushBatchObsFunc,
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
				h.pushObsFunc(ctx, o)
			case m := <-h.batchObsvC:
				o := m.Msg
				h.guardian.Ping(ctx)
				h.metrics.IncBatchObservationTotal()
				h.pushBatchObsFunc(ctx, o)
			}
		}
	}()
}
