package governor

import (
	"time"

	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type GovernorVaasResponse struct {
	VaaID          string      `json:"vaaId"`
	ChainID        vaa.ChainID `json:"chainId"`
	EmitterAddress string      `json:"emitterAddress"`
	Sequence       string      `json:"sequence"`
	TxHash         string      `json:"txHash"`
	ReleaseTime    time.Time   `json:"releaseTime"`
	Amount         uint64      `json:"amount"`
	Status         string      `json:"status"`
}
