package stats

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// NativeTokenTransferTopAddress represents the top address of native token transfer
type NativeTokenTransferTopAddress struct {
	FromAddress string          `json:"fromAddress"`
	Value       decimal.Decimal `json:"value"`
}

// NativeTokenTransferTopHolder
type NativeTokenTransferTopHolder struct {
	Address string          `json:"address"`
	ChainID sdk.ChainID     `json:"chain"`
	Value   decimal.Decimal `json:"value"`
}

type cachedResult[T any] struct {
	Timestamp time.Time `json:"timestamp"`
	Result    T         `json:"result"`
}

func (c cachedResult[T]) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)
}
