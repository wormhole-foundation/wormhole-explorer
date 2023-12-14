package processor

import (
	"context"
	"strconv"
	"strings"
	"sync"

	"github.com/certusone/wormhole/node/pkg/common"
	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
	eth_common "github.com/ethereum/go-ethereum/common"
	crypto2 "github.com/ethereum/go-ethereum/crypto"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/fly/storage"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type observationGossipConsumer struct {
	signedObsCh chan *gossipv1.SignedObservation
	repository  *storage.Repository
	gst         *common.GuardianSetState
	environment string
	workerSize  int
	metrics     metrics.Metrics
	wgBlock     sync.WaitGroup
	logger      *zap.Logger
}

// NewObservationGossipConsumer creates a new processor instances.
func NewObservationGossipConsumer(
	repository *storage.Repository,
	gst *common.GuardianSetState,
	environment string,
	channelSize int,
	workerSize int,
	metrics metrics.Metrics,
	logger *zap.Logger,
) *observationGossipConsumer {
	return &observationGossipConsumer{
		repository:  repository,
		gst:         gst,
		environment: environment,
		workerSize:  workerSize,
		metrics:     metrics,
		logger:      logger,
		signedObsCh: make(chan *gossipv1.SignedObservation, channelSize),
	}
}

// Start starts the processor.
func (c *observationGossipConsumer) Start(ctx context.Context) error {
	for i := 0; i < c.workerSize; i++ {
		c.wgBlock.Add(1)
		go c.run(ctx)
	}
	return nil
}

// Push pushes a new observation to the processor.
func (c *observationGossipConsumer) Push(ctx context.Context, o *gossipv1.SignedObservation) error {
	c.signedObsCh <- o
	return nil
}

// Close closes all consumer resources.
func (c *observationGossipConsumer) Close() {
	close(c.signedObsCh)
	c.wgBlock.Wait()
}

func (c *observationGossipConsumer) run(ctx context.Context) error {
	defer c.wgBlock.Done()
	for {
		select {
		case <-ctx.Done():
			return nil
		case o := <-c.signedObsCh:
			c.process(ctx, o)
		}
	}
}

func (c *observationGossipConsumer) process(ctx context.Context, o *gossipv1.SignedObservation) {
	ok := verifyObservation(c.logger, o, c.gst.Get())
	if !ok {
		c.logger.Error("Could not verify observation", zap.String("id", o.MessageId))
		return
	}

	// get chainID from observationID.
	chainID, err := getObservationChainID(c.logger, o)
	if err != nil {
		c.logger.Error("Error getting chainID", zap.Error(err))
		return
	}
	c.metrics.IncObservationFromGossipNetwork(chainID)

	// apply filter observations by env.
	if filterObservationByEnv(o, c.environment) {
		return
	}

	c.metrics.IncObservationUnfiltered(chainID)

	err = c.repository.UpsertObservation(o)
	if err != nil {
		c.logger.Error("Error inserting observation", zap.Error(err))
	}
}

func verifyObservation(logger *zap.Logger, obs *gossipv1.SignedObservation, gs *common.GuardianSet) bool {
	pk, err := crypto2.Ecrecover(obs.GetHash(), obs.GetSignature())
	if err != nil {
		return false
	}

	theirAddr := eth_common.BytesToAddress(obs.GetAddr())
	signerAddr := eth_common.BytesToAddress(crypto2.Keccak256(pk[1:])[12:])
	if theirAddr != signerAddr {
		logger.Error("error validating observation, signer addr and addr don't match",
			zap.String("id", obs.MessageId),
			zap.String("obs_addr", theirAddr.Hex()),
			zap.String("signer_addr", signerAddr.Hex()),
		)
		return false
	}

	_, isFromGuardian := gs.KeyIndex(theirAddr)
	if !isFromGuardian {
		logger.Error("error validating observation, signer not in guardian set",
			zap.String("id", obs.MessageId),
			zap.String("obs_addr", theirAddr.Hex()),
		)
	}
	return isFromGuardian
}

func getObservationChainID(logger *zap.Logger, obs *gossipv1.SignedObservation) (sdk.ChainID, error) {
	vaaID := strings.Split(obs.MessageId, "/")
	chainIDStr := vaaID[0]
	chainID, err := strconv.ParseUint(chainIDStr, 10, 16)
	if err != nil {
		logger.Error("Error parsing chainId", zap.Error(err))
		return 0, err
	}
	return sdk.ChainID(chainID), nil
}

// filterObservation filter observation by enviroment.
func filterObservationByEnv(o *gossipv1.SignedObservation, enviroment string) bool {
	if enviroment == domain.P2pTestNet {
		// filter pyth message in testnet gossip network (for solana and pyth chain).
		if strings.Contains((o.GetMessageId()), "1/f346195ac02f37d60d4db8ffa6ef74cb1be3550047543a4a9ee9acf4d78697b0") ||
			strings.HasPrefix("26/", o.GetMessageId()) {
			return true
		}
	}
	// filter pyth message in mainnet gossip network (for pyth chain).
	if enviroment == domain.P2pMainNet && strings.HasPrefix("26/", o.GetMessageId()) {
		return true
	}
	return false
}
