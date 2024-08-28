package stats

import (
	"context"

	"github.com/go-resty/resty/v2"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	"go.uber.org/zap"
)

type HolderRepository struct {
	client        *resty.Client
	arkhamUrl     string
	artkhamApiKey string
	solanaUrl     string
	solanaApiKey  string
	cache         cache.Cache
	log           *zap.Logger
}

func NewHolderRepository(client *resty.Client, arkhamUrl, arkhamApiKey, solanaUrl, solanaApiKey string, cache cache.Cache, log *zap.Logger) *HolderRepository {
	return &HolderRepository{
		client:        resty.New(),
		arkhamUrl:     arkhamUrl,
		artkhamApiKey: arkhamApiKey,
		solanaUrl:     solanaUrl,
		solanaApiKey:  solanaApiKey,
		cache:         cache,
		log:           log,
	}
}

func (r *HolderRepository) LoadNativeTokenTransferTopHolder(ctx context.Context, symbol string) error {
	return nil
}

func (r *HolderRepository) GetNativeTokenTransferTopHolder(ctx context.Context, symbol string) ([]NativeTokenTransferTopHolder, error) {
	return nil, nil
}
