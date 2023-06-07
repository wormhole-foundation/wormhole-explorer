package transactions

import "time"

// TransactionOverview is a brief description of a transaction (e.g. ID, txHash, status, etc.).
type TransactionOverview struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}

// ListTransactionsResponse is the "200 OK" response model for `GET /api/v1/transactions`.
type ListTransactionsResponse struct {
	Transactions []TransactionOverview `json:"transactions"`
}
