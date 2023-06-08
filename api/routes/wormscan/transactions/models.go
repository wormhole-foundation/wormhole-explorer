package transactions

import (
	"time"

	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// TransactionOverview is a brief description of a transaction (e.g. ID, txHash, status, etc.).
type TransactionOverview struct {
	ID                 string      `json:"id"`
	Timestamp          time.Time   `json:"timestamp"`
	TxHash             string      `json:"txHash"`
	OriginChain        sdk.ChainID `json:"originChain"`
	DestinationAddress string      `json:"destinationAddress,omitempty"`
	DestinationChain   sdk.ChainID `json:"destinationChain,omitempty"`
	TokenAmount        string      `json:"tokenAmount,omitempty"`
	UsdAmount          string      `json:"usdAmount,omitempty"`
	Symbol             string      `json:"symbol,omitempty"`
}

// ListTransactionsResponse is the "200 OK" response model for `GET /api/v1/transactions`.
type ListTransactionsResponse struct {
	Transactions []TransactionOverview `json:"transactions"`
}
