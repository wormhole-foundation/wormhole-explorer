package transactions

import (
	"context"
	"fmt"
	"strconv"

	"github.com/wormhole-foundation/wormhole-explorer/api/internal/config"
	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"go.uber.org/zap"
)

type totalPythResult struct {
	ID       string `data:"id"`
	Sequence uint64 `data:"sequence"`
}

type PostgresRepository struct {
	p2pNetwork string
	db         *db.DB
	logger     *zap.Logger
}

func NewPostgresRepository(p2pNetwork string, db *db.DB, logger *zap.Logger) *PostgresRepository {
	return &PostgresRepository{
		p2pNetwork: p2pNetwork,
		db:         db,
		logger:     logger.With(zap.String("module", "PostgresTransactionsRepository")),
	}
}

// getTotalPythMessage returns the last sequence for the pyth emitter address
func (r *PostgresRepository) getTotalPythMessage(ctx context.Context) (string, error) {
	if r.p2pNetwork != config.P2pMainNet {
		return "0", nil

	}
	query := `
	SELECT id, sequence FROM wormholescan.wh_attestation_vaas_pythnet 
    WHERE emitter_address = $1
    ORDER BY "timestamp" DESC 
    LIMIT 1`

	var result totalPythResult
	err := r.db.SelectOne(ctx, &result, query, pythEmitterAddr)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		r.logger.Error("failed to execute query", zap.Error(err), zap.String("query", query))
		return "", err
	}

	return strconv.FormatUint(result.Sequence, 10), nil
}
