package txhash

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

var ErrTxHashNotFound = errors.New("tx hash not found")

type TxHash struct {
	ChainID  sdk.ChainID
	Emitter  string
	Sequence string
	TxHash   string
}

type TxHashStore interface {
	Get(ctx context.Context, vaaID string) (*string, error)
	Set(ctx context.Context, vaaID string, txHash TxHash) error
	SetObservation(ctx context.Context, o *gossipv1.SignedObservation) error
	GetName() string
}

func CreateTxHash(logger *zap.Logger, o *gossipv1.SignedObservation) (*TxHash, error) {
	vaaID := strings.Split(o.MessageId, "/")
	chainIDStr, emitter, sequenceStr := vaaID[0], vaaID[1], vaaID[2]

	chainID, err := strconv.ParseUint(chainIDStr, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("error parsing chainId: %w", err)
	}

	txHash, err := domain.EncodeTrxHashByChainID(sdk.ChainID(chainID), o.GetTxHash())
	if err != nil {
		logger.Warn("Error encoding tx hash",
			zap.Uint64("chainId", chainID),
			zap.ByteString("txHash", o.GetTxHash()),
			zap.Error(err))
	}

	return &TxHash{
		ChainID:  sdk.ChainID(chainID),
		Emitter:  emitter,
		Sequence: sequenceStr,
		TxHash:   txHash,
	}, nil
}
