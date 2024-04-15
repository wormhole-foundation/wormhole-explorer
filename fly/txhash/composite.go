package txhash

import (
	"context"
	"strconv"
	"strings"

	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
	"github.com/hashicorp/go-multierror"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/metrics"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type composite struct {
	hashStores []TxHashStore
	metrics    metrics.Metrics
	logger     *zap.Logger
}

func NewComposite(
	hashStores []TxHashStore,
	metrics metrics.Metrics,
	logger *zap.Logger) *composite {
	return &composite{
		hashStores: hashStores,
		metrics:    metrics,
		logger:     logger,
	}
}

func (t *composite) Set(ctx context.Context, uniqueVaaID string, txHash TxHash) error {
	var result multierror.Error
	for _, store := range t.hashStores {
		if err := store.Set(ctx, uniqueVaaID, txHash); err != nil {
			t.logger.Error("Error setting tx hash",
				zap.String("vaaId", uniqueVaaID),
				zap.String("store", store.GetName()),
				zap.Error(err))
			result.Errors = append(result.Errors, err)
		}
	}
	return result.ErrorOrNil()
}

func (t *composite) SetObservation(ctx context.Context, o *gossipv1.SignedObservation) error {
	vaaID := strings.Split(o.MessageId, "/")
	chainIDStr, emitter, sequenceStr := vaaID[0], vaaID[1], vaaID[2]

	chainID, err := strconv.ParseUint(chainIDStr, 10, 16)
	if err != nil {
		t.logger.Error("Error parsing chainId", zap.Error(err))
		return err
	}

	txHash, err := domain.EncodeTrxHashByChainID(vaa.ChainID(chainID), o.GetTxHash())
	if err != nil {
		t.logger.Warn("Error encoding tx hash",
			zap.Uint64("chainId", chainID),
			zap.ByteString("txHash", o.GetTxHash()),
			zap.Error(err))
	}

	vaaTxHash := TxHash{
		ChainID:  vaa.ChainID(chainID),
		Emitter:  emitter,
		Sequence: sequenceStr,
		TxHash:   txHash,
	}
	uniqueVaaID := domain.CreateUniqueVaaIDByObservation(o)
	return t.Set(ctx, uniqueVaaID, vaaTxHash)
}

func (t *composite) Get(ctx context.Context, uniqueVaaID string) (*string, error) {
	log := t.logger.With(zap.String("vaaId", uniqueVaaID))
	for _, store := range t.hashStores {
		txHash, err := store.Get(ctx, uniqueVaaID)
		if err == nil {
			t.metrics.IncFoundTxHash(store.GetName())
			return txHash, nil
		}
		if err != ErrTxHashNotFound {
			log.Error("Error getting tx hash", zap.String("store", store.GetName()), zap.Error(err))
		} else {
			log.Info("Cannot get txHash", zap.String("store", store.GetName()))
		}
		t.metrics.IncNotFoundTxHash(store.GetName())
	}
	return nil, ErrTxHashNotFound
}

func (t *composite) GetName() string {
	return "multi"
}
