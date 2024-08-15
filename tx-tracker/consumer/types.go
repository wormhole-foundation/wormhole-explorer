package consumer

import (
	"context"
)

// Repository is the interface that wraps the basic methods to interact with the database.
type Repository interface {
	AlreadyProcessed(ctx context.Context, vaaID string, vaaDigest string) (bool, error)
	GetVaaIdTxHash(ctx context.Context, vaaID string, vaaDigest string) (*VaaIdTxHash, error)
	UpsertTargetTx(ctx context.Context, globalTx *TargetTxUpdate) error
	GetTxStatus(ctx context.Context, targetTxUpdate *TargetTxUpdate) (string, error)
	FindSourceTxById(ctx context.Context, id string) (*SourceTxDoc, error)
	UpsertOriginTx(ctx context.Context, originTx, nestedTx *UpsertOriginTxParams) error
}
