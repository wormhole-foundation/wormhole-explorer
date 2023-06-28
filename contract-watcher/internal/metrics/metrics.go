package metrics

import (
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

const serviceName = "wormscan-contract-watcher"

type Metrics interface {
	SetLastBlock(chain sdk.ChainID, block uint64)
	SetCurrentBlock(chain sdk.ChainID, block uint64)
	IncDestinationTrxSaved(chain sdk.ChainID)
	IncRpcRequest(client string, method string, statusCode int)
}
