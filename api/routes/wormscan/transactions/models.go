package transactions

import (
	"time"

	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type TxStatus string

const (
	TxStatusOngoing   TxStatus = "ongoing"
	TxStatusCompleted TxStatus = "completed"
)

// TransactionOverview is a brief description of a transaction (e.g. ID, txHash, status, etc.).
type TransactionOverview struct {
	ID            string      `json:"id"`
	Timestamp     time.Time   `json:"timestamp"`
	TxHash        string      `json:"txHash,omitempty"`
	OriginAddress string      `json:"originAddress,omitempty"`
	OriginChain   sdk.ChainID `json:"originChain"`
	// EmitterAddress contains the VAA's emitter address, encoded in hex.
	EmitterAddress string `json:"emitterAddress"`
	// EmitterNativeAddress contains the VAA's emitter address in the emitter chain's native format.
	EmitterNativeAddress string      `json:"emitterNativeAddress,omitempty"`
	DestinationAddress   string      `json:"destinationAddress,omitempty"`
	DestinationChain     sdk.ChainID `json:"destinationChain,omitempty"`
	TokenAmount          string      `json:"tokenAmount,omitempty"`
	UsdAmount            string      `json:"usdAmount,omitempty"`
	Symbol               string      `json:"symbol,omitempty"`
	Status               TxStatus    `json:"status"`
}

// ListTransactionsResponse is the "200 OK" response model for `GET /api/v1/transactions`.
type ListTransactionsResponse struct {
	Transactions []*TransactionOverview `json:"transactions"`
}
