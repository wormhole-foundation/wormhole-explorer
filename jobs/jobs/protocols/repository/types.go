package repository

import (
	"context"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/protocols/internal/commons"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type ProtocolRepository interface {
	GetActivity(ctx context.Context, from, to time.Time) (ProtocolActivity, error)
	GetStats(ctx context.Context) (Stats, error)
	ProtocolName() string
}

type ProtocolActivity struct {
	TotalValueSecure      float64    `json:"total_value_secure"`
	TotalValueTransferred float64    `json:"total_value_transferred"`
	Volume                float64    `json:"volume"`
	TotalMessages         uint64     `json:"total_messages"`
	Activities            []Activity `json:"activity"`
}

type Stats struct {
	TotalValueLocked float64 `json:"total_value_locked"`
	TotalMessages    uint64  `json:"total_messages"`
	Volume           float64 `json:"volume"`
}

type Activity struct {
	EmitterChainID     uint64  `json:"emitter_chain_id"`
	DestinationChainID uint64  `json:"destination_chain_id"`
	Txs                uint64  `json:"txs"`
	TotalUSD           float64 `json:"total_usd"`
}

// ProtocolsRepositoryFactory RestClient Factory to create the right client for each protocol.
var ProtocolsRepositoryFactory = map[string]func(url string, logger *zap.Logger) ProtocolRepository{

	commons.MayanProtocol: func(baseURL string, logger *zap.Logger) ProtocolRepository {
		return NewMayanRestClient(baseURL, logger, &http.Client{})
	},

	commons.AllBridgeProtocol: func(baseURL string, logger *zap.Logger) ProtocolRepository {
		return NewAllBridgeRestClient(baseURL, logger, &http.Client{})
	},
}
