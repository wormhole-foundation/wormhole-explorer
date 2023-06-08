package transactions

import (
	"time"

	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// TransactionOverview is a brief description of a transaction (e.g. ID, txHash, status, etc.).
type TransactionOverview struct {
	ID                 string      `json:"id"`
	Timestamp          time.Time   `json:"timestamp"`
	DestinationAddress string      `json:"destinationAddress"`
	DestinationChain   sdk.ChainID `json:"destinationChain"`
	TokenAmount        string      `json:"tokenAmount"`
	UsdAmount          string      `json:"usdAmount"`
	Symbol             string      `json:"symbol"`
}

// ListTransactionsResponse is the "200 OK" response model for `GET /api/v1/transactions`.
type ListTransactionsResponse struct {
	Transactions []TransactionOverview `json:"transactions"`
}
