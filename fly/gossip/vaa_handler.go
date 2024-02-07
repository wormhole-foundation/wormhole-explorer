package gossip

import (
	"context"
	"strings"

	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/fly/config"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/health"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/fly/processor"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type vaaHandler struct {
	p2pNetworkConfig *config.P2pNetworkConfig
	metrics          metrics.Metrics
	signedInC        chan *gossipv1.SignedVAAWithQuorum
	vaaHandlerFunc   processor.VAAPushFunc
	guardian         *health.GuardianCheck
	logger           *zap.Logger
}

func NewVaaHandler(
	p2pNetworkConfig *config.P2pNetworkConfig,
	metrics metrics.Metrics,
	signedInC chan *gossipv1.SignedVAAWithQuorum,
	vaaHandlerFunc processor.VAAPushFunc,
	guardian *health.GuardianCheck,
	logger *zap.Logger,
) *vaaHandler {
	return &vaaHandler{
		p2pNetworkConfig: p2pNetworkConfig,
		metrics:          metrics,
		signedInC:        signedInC,
		vaaHandlerFunc:   vaaHandlerFunc,
		guardian:         guardian,
		logger:           logger,
	}
}

func (h *vaaHandler) Start(rootCtx context.Context) {
	go func() {
		for {
			select {
			case <-rootCtx.Done():
				return
			case sVaa := <-h.signedInC:
				h.guardian.Ping(rootCtx)
				h.metrics.IncVaaTotal()
				vaa, err := sdk.Unmarshal(sVaa.Vaa)
				if err != nil {
					h.logger.Error("Error unmarshalling vaa", zap.Error(err))
					continue
				}

				h.metrics.IncVaaFromGossipNetwork(vaa.EmitterChain)
				// apply filter observations by env.
				if filterVaasByEnv(vaa, h.p2pNetworkConfig.Enviroment) {
					continue
				}

				// Push an incoming VAA to be processed
				if err := h.vaaHandlerFunc(rootCtx, vaa, sVaa.Vaa); err != nil {
					h.logger.Error("Error inserting vaa", zap.Error(err))
				}
			}
		}
	}()
}

// filterVaasByEnv filter vaa by enviroment.
func filterVaasByEnv(v *sdk.VAA, enviroment string) bool {
	if enviroment == domain.P2pTestNet {
		vaaFromSolana := v.EmitterChain == sdk.ChainIDSolana
		addressToFilter := strings.ToLower(v.EmitterAddress.String()) == "f346195ac02f37d60d4db8ffa6ef74cb1be3550047543a4a9ee9acf4d78697b0"
		isPyth := v.EmitterChain == sdk.ChainIDPythNet
		if (vaaFromSolana && addressToFilter) || isPyth {
			return true
		}
	}
	return false
}
