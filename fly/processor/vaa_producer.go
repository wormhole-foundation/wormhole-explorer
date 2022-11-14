package processor

import (
	"context"

	"github.com/certusone/wormhole/node/pkg/common"
	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type VAAPublish func(context.Context, *vaa.VAA, []byte) error

type VAAProducer struct {
	gst            *common.GuardianSetState
	nonPythPublish VAAPublish
	pythPublish    VAAPublish
	logger         *zap.Logger
}

func NewVAAProducer(
	gst *common.GuardianSetState,
	nonPythPublish VAAPublish,
	pythPublish VAAPublish,
	logger *zap.Logger) *VAAProducer {
	return &VAAProducer{
		gst:            gst,
		nonPythPublish: nonPythPublish,
		pythPublish:    pythPublish,
		logger:         logger,
	}
}

func (p *VAAProducer) Push(ctx context.Context, sVaa *gossipv1.SignedVAAWithQuorum) error {
	v, err := vaa.Unmarshal(sVaa.Vaa)
	if err != nil {
		p.logger.Error("Error unmarshalling vaa", zap.Error(err))
		return err
	}
	if err := v.Verify(p.gst.Get().Keys); err != nil {
		p.logger.Error("Received invalid vaa", zap.String("id", v.MessageID()))
		return err
	}
	if vaa.ChainIDPythNet == v.EmitterChain {
		err = p.pythPublish(ctx, v, sVaa.Vaa)
	} else {
		err = p.nonPythPublish(ctx, v, sVaa.Vaa)
	}
	if err != nil {
		p.logger.Error("Error inserting vaa in store", zap.Error(err))
		return err
	}
	return nil
}
