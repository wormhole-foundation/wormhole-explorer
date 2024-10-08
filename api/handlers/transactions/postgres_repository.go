package transactions

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/api/internal/config"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
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

type operationTxResult struct {
	ChainID           uint16         `db:"chain_id"`
	TxHash            string         `db:"tx_hash"`
	Type              string         `db:"type"`
	CreatedAt         time.Time      `db:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at"`
	AttestationVaasID string         `db:"attestation_vaas_id"`
	MessageID         string         `db:"message_id"`
	Status            *string        `db:"status"`
	FromAddress       *string        `db:"from_address"`
	ToAddress         *string        `db:"to_address"`
	BlockNumber       *string        `db:"block_number"`
	BlockchainMethod  *string        `db:"blockchain_method"`
	FeeDetail         map[string]any `db:"fee_detail"`
	Timestamp         time.Time      `db:"timestamp"`
	RPCResponse       map[string]any `db:"rpc_response"`
}

// FindGlobalTransactionByID returns a global transaction by its ID.
func (r *PostgresRepository) FindGlobalTransactionByID(
	ctx context.Context,
	q *GlobalTransactionQuery,
) (*GlobalTransactionDoc, error) {

	query := `SELECT chain_id, tx_hash, "type", created_at, updated_at, attestation_vaas_id, message_id, status, 
	from_address, to_address, block_number, blockchain_method, fee_detail, timestamp, rpc_response 
	FROM wormholescan.wh_operation_transactions 
	WHERE message_id = $1
	`

	var operationTxs []*operationTxResult
	err := r.db.Select(ctx, &operationTxs, query, q.id)
	if err != nil {
		r.logger.Error("failed to execute query", zap.Error(err), zap.String("query", query))
		return nil, err
	}

	globalTransaction, err := createGlobalTransaction(operationTxs)
	if err != nil {
		r.logger.Error("failed to create global transaction", zap.Error(err))
		return nil, err
	}

	return globalTransaction, nil
}

func createGlobalTransaction(operationTxs []*operationTxResult) (*GlobalTransactionDoc, error) {
	if len(operationTxs) == 0 {
		return nil, errors.ErrNotFound
	}

	// check all the operation txs have the same message id
	messageID := operationTxs[0].MessageID
	for _, tx := range operationTxs {
		if tx.MessageID != messageID {
			return nil, fmt.Errorf("operation txs have different message ids %s and %s", messageID, tx.MessageID)
		}
	}

	// create origin and destination txs
	originTx := createOriginTx(operationTxs)
	destinationTx := createDestinationTx(operationTxs)

	// create global transaction
	return &GlobalTransactionDoc{
		ID:            messageID,
		OriginTx:      originTx,
		DestinationTx: destinationTx,
	}, nil
}

func createOriginTx(operationTxs []*operationTxResult) *OriginTx {
	var sourceTx *operationTxResult
	var nestedSourceTx *operationTxResult
	for _, tx := range operationTxs {
		if tx.Type == "source-tx" {
			sourceTx = tx
		}
		if tx.Type == "nested-source-tx" {
			nestedSourceTx = tx
		}
	}

	// denormalize tx hash for compatibility reasons.
	denormalizedTxHash := domain.DenormalizeTxHashByChainId(
		sdk.ChainID(sourceTx.ChainID), sourceTx.TxHash)

	// denormalize from address for compatibility reasons.
	var denormalizedFromAddress string
	if sourceTx.FromAddress != nil {
		denormalizedFromAddress = domain.DenormalizeAddressByChainId(
			sdk.ChainID(sourceTx.ChainID), *sourceTx.FromAddress)
	}

	originTx := &OriginTx{
		TxHash: denormalizedTxHash,
		From:   denormalizedFromAddress,
		Status: *sourceTx.Status,
	}

	var attribute *AttributeDoc
	if nestedSourceTx != nil {

		var denormalizedNestedTxHash string
		if nestedSourceTx.TxHash != "" {
			// denormalize tx hash for compatibility reasons.
			denormalizedNestedTxHash = domain.DenormalizeTxHashByChainId(
				sdk.ChainID(nestedSourceTx.ChainID), nestedSourceTx.TxHash)
		}
		values := map[string]any{
			"originChainId": nestedSourceTx.ChainID,
			"originTxHash":  denormalizedNestedTxHash,
		}
		if nestedSourceTx.FromAddress != nil {
			// denormalize from address for compatibility reasons.
			denormalizedOriginAddress := domain.DenormalizeAddressByChainId(
				sdk.ChainID(nestedSourceTx.ChainID), *nestedSourceTx.FromAddress)
			values["originAddress"] = denormalizedOriginAddress
		}

		attribute = &AttributeDoc{
			Type:  "wormchain-gateway",
			Value: values,
		}
	}
	originTx.Attribute = attribute
	return originTx
}

func createDestinationTx(operationTxs []*operationTxResult) *DestinationTx {
	var destinationTx *operationTxResult
	for _, tx := range operationTxs {
		if tx.Type == "target" {
			destinationTx = tx
		}
	}

	if destinationTx == nil {
		return nil
	}

	var blockNumber string
	if destinationTx.BlockNumber != nil {
		blockNumber = *destinationTx.BlockNumber
	}

	var timestamp *time.Time
	if !destinationTx.Timestamp.IsZero() {
		timestamp = &destinationTx.Timestamp
	}

	var updatedAt *time.Time
	if !destinationTx.UpdatedAt.IsZero() {
		updatedAt = &destinationTx.UpdatedAt
	}

	// denormalize tx hash for compatibility reasons.
	denormalizedTxHash := domain.DenormalizeTxHashByChainId(
		sdk.ChainID(destinationTx.ChainID), destinationTx.TxHash)

	// denormalize from address for compatibility reasons.
	var denormalizedFromAddress string
	if destinationTx.FromAddress != nil {
		denormalizedFromAddress = domain.DenormalizeAddressByChainId(
			sdk.ChainID(destinationTx.ChainID), *destinationTx.FromAddress)
	}

	// denormalize to address for compatibility reasons.
	var denormalizedToAddress string
	if destinationTx.ToAddress != nil {
		denormalizedToAddress = domain.DenormalizeAddressByChainId(
			sdk.ChainID(destinationTx.ChainID), *destinationTx.ToAddress)
	}

	return &DestinationTx{
		ChainID:     sdk.ChainID(destinationTx.ChainID),
		Status:      *destinationTx.Status,
		Method:      *destinationTx.BlockchainMethod,
		TxHash:      denormalizedTxHash,
		From:        denormalizedFromAddress,
		To:          denormalizedToAddress,
		BlockNumber: blockNumber,
		Timestamp:   timestamp,
		UpdatedAt:   updatedAt,
	}
}
