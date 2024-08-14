package consumer

import (
	"context"
	"strconv"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type PostgreSQLRepository struct {
	dbClient *db.DB
}

func NewPostgreSQLRepository(postreSQLClient *db.DB) *PostgreSQLRepository {
	return &PostgreSQLRepository{
		dbClient: postreSQLClient,
	}
}

func (p *PostgreSQLRepository) UpsertOriginTx(ctx context.Context, originTx, nested *UpsertOriginTxParams) error {
	if err := p.upsertOriginTx(ctx, originTx); err != nil {
		return err
	}

	if nested != nil {
		if err := p.upsertOriginTx(ctx, nested); err != nil {
			return err
		}
	}

	return nil
}

func (p *PostgreSQLRepository) upsertOriginTx(ctx context.Context, params *UpsertOriginTxParams) error {

	query := `
		INSERT INTO wormholescan.wh_operation_transactions 
		(chain_id, tx_hash, type, created_at, updated_at, attestation_vaas_id, vaa_id, status, from_address, to_address, block_number, blockchain_method, fee, raw_fee, timestamp, rpc_response)  
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (chain_id, tx_hash) DO UPDATE
		SET 
			type = EXCLUDED.type,
			created_at = wormholescan.wh_operation_transactions.created_at,
			updated_at = EXCLUDED.updated_at,
			attestation_vaas_id = EXCLUDED.attestation_vaas_id,
			vaa_id = EXCLUDED.vaa_id,
			status = EXCLUDED.status,
			from_address = COALESCE(EXCLUDED.from_address, wormholescan.wh_operation_transactions.from_address),
			to_address = COALESCE(EXCLUDED.to_address, wormholescan.wh_operation_transactions.to_address),
			block_number = COALESCE(EXCLUDED.block_number, wormholescan.wh_operation_transactions.block_number),
			blockchain_method = COALESCE(EXCLUDED.blockchain_method, wormholescan.wh_operation_transactions.blockchain_method),
			fee = COALESCE(EXCLUDED.fee, wormholescan.wh_operation_transactions.fee),
			raw_fee = COALESCE(EXCLUDED.raw_fee, wormholescan.wh_operation_transactions.raw_fee),
			timestamp = EXCLUDED.timestamp,
			rpc_response = COALESCE(EXCLUDED.rpc_response, wormholescan.wh_operation_transactions.rpc_response)
		`

	var fee *string
	var rawFee map[string]string
	var from, to, rpcResponse, nativeTxHash *string
	var blockNumber *uint64
	if params.TxDetail != nil {
		from = &params.TxDetail.From
		to = &params.TxDetail.To
		nativeTxHash = &params.TxDetail.NativeTxHash
		if params.TxDetail.FeeDetail != nil {
			fee = &params.TxDetail.FeeDetail.Fee
			rawFee = params.TxDetail.FeeDetail.RawFee
		}

		bn, errBn := strconv.ParseUint(params.TxDetail.BlockNumber, 10, 64)
		if errBn == nil {
			blockNumber = &bn
		}

	}

	_, err := p.dbClient.Exec(ctx, query,
		params.ChainId,
		nativeTxHash,
		"source-tx",      // type
		time.Now(),       // created_at
		time.Now(),       // updated_at
		params.Id,        // attestation_vaas_id
		params.VaaId,     // vaa_id
		params.TxStatus,  // status
		from,             // from_address
		to,               // to_address
		blockNumber,      // block_number
		nil,              // blockchain_method: only applies for incoming targetTx from blockchain-watcher
		fee,              // fee
		rawFee,           // raw_fee
		params.Timestamp, // timestamp
		rpcResponse,      // rpc_response
	)

	return err
}

func (p *PostgreSQLRepository) UpsertTargetTx(ctx context.Context, params *TargetTxUpdate) error {
	query := `
		INSERT INTO wormholescan.wh_operation_transactions 
		(chain_id, tx_hash, type, created_at, updated_at, attestation_vaas_id, vaa_id, status, from_address, to_address, block_number, blockchain_method, fee, raw_fee, timestamp, rpc_response)  
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (chain_id, tx_hash) DO UPDATE
		SET 
			type = EXCLUDED.type,
			created_at = wormholescan.wh_operation_transactions.created_at,
			updated_at = EXCLUDED.updated_at,
			attestation_vaas_id = EXCLUDED.attestation_vaas_id,
			vaa_id = EXCLUDED.vaa_id,
			status = COALESCE(EXCLUDED.status, wormholescan.wh_operation_transactions.status),
			from_address = COALESCE(EXCLUDED.from_address, wormholescan.wh_operation_transactions.from_address),
			to_address = COALESCE(EXCLUDED.to_address, wormholescan.wh_operation_transactions.to_address),
			block_number = COALESCE(EXCLUDED.block_number, wormholescan.wh_operation_transactions.block_number),
			blockchain_method = COALESCE(EXCLUDED.blockchain_method, wormholescan.wh_operation_transactions.blockchain_method),
			fee = COALESCE(EXCLUDED.fee, wormholescan.wh_operation_transactions.fee),
			raw_fee = COALESCE(EXCLUDED.raw_fee, wormholescan.wh_operation_transactions.raw_fee),
			timestamp = EXCLUDED.timestamp,
			rpc_response = COALESCE(EXCLUDED.rpc_response, wormholescan.wh_operation_transactions.rpc_response)
		`

	var fee *string
	var rawFee map[string]string
	var from, to, blockchainMethod, status, txHash *string
	var blockNumber *uint64
	var timestamp, updatedAt *time.Time
	var chainID *vaa.ChainID
	if params.Destination != nil {
		from = &params.Destination.From
		to = &params.Destination.To
		blockchainMethod = &params.Destination.Method
		status = &params.Destination.Status
		txHash = &params.Destination.TxHash
		chainID = &params.Destination.ChainID
		timestamp = params.Destination.Timestamp
		updatedAt = params.Destination.UpdatedAt
		if params.Destination.FeeDetail != nil {
			fee = &params.Destination.FeeDetail.Fee
			rawFee = params.Destination.FeeDetail.RawFee
		}

		bn, errBn := strconv.ParseUint(params.Destination.BlockNumber, 10, 64)
		if errBn == nil {
			blockNumber = &bn
		}

	}
	_, err := p.dbClient.Exec(ctx, query,
		chainID,
		txHash,
		"target-tx",      // type
		time.Now(),       // created_at
		updatedAt,        // updated_at
		params.ID,        // attestation_vaas_id
		params.VaaID,     // vaa_id
		status,           // status
		from,             // from_address
		to,               // to_address
		blockNumber,      // block_number
		blockchainMethod, // blockchain_method
		fee,              // fee
		rawFee,           // raw_fee
		timestamp,        // timestamp
		nil,              // rpc_response
	)

	return err
}

func (p *PostgreSQLRepository) GetTxStatus(ctx context.Context, targetTxUpdate *TargetTxUpdate) (string, error) {
	var status string
	err := p.dbClient.SelectOne(ctx, &status, `SELECT status FROM wormholescan.wh_operation_transactions WHERE chain_id = $1 AND tx_hash = $2`, targetTxUpdate.Destination.ChainID, targetTxUpdate.Destination.TxHash)
	return status, err
}

func (p *PostgreSQLRepository) AlreadyProcessed(ctx context.Context, vaaId string, digest string) (bool, error) {
	var count int
	err := p.dbClient.SelectOne(ctx, &count, `SELECT COUNT(*) FROM wormholescan.wh_operation_transactions_processed WHERE id = $1`, digest)
	return count > 0, err
}

func (p *PostgreSQLRepository) RegisterProcessedVaa(ctx context.Context, vaaDigest, vaaId string) error {
	now := time.Now()
	_, err := p.dbClient.Exec(ctx,
		`INSERT INTO wormholescan.wh_operation_transactions_processed (id,vaa_id,processed,created_at,updated_at)
			VALUES ($1,$2,true,$3,$4)`, vaaDigest, vaaId, now, now)
	return err
}

// GetVaaIdTxHash returns the VaaIdTxHash for the given id. this dummy implementation is added in postgres repository
// to support the Repository interface. Remove this method after migrations.
func (p *PostgreSQLRepository) GetVaaIdTxHash(ctx context.Context, vaaID, vaaDigest string) (*VaaIdTxHash, error) {
	var txHash string
	err := p.dbClient.SelectOne(ctx, &txHash, "SELECT tx_hash FROM wormholescan.wh_observations WHERE wh_observations.hash = $1", vaaDigest)
	if err != nil {
		return nil, err
	}
	return &VaaIdTxHash{TxHash: txHash}, nil
}

// FindSourceTxById returns the source tx by id. this dummy implementation is added in postgres repository
// to support the Repository interface. Remove this method after migrations.
func (p *PostgreSQLRepository) FindSourceTxById(ctx context.Context, id string) (*SourceTxDoc, error) {

	var sourceTx struct {
		ID       string `db:"vaa_id"`
		TxHash   string `db:"tx_hash"`
		ChainID  uint16 `db:"emitter_chain_id"`
		Status   string `db:"status"`
		FromAddr string `db:"from_address"`
	}

	query := `
	SELECT o.vaa_id, .o.tx_hash, v.emitter_chain_id, o.status, o.from_address
	FROM wormholescan.wh_attestation_vaas as v
	INNER JOIN wormholescan.wh_operation_transactions as o ON o.id = v.id 
	WHERE v.vaa_id = $1 and v.active = true and o.type = 'source-tx'
	`

	err := p.dbClient.SelectOne(ctx, &sourceTx, query, id)
	if err != nil {
		return nil, err
	}

	return &SourceTxDoc{
		ID: sourceTx.ID,
		OriginTx: &struct {
			ChainID      int    `bson:"chainId"`
			Status       string `bson:"status"`
			Processed    bool   `bson:"processed"`
			NativeTxHash string `bson:"nativeTxHash"`
			From         string `bson:"from"`
		}{
			ChainID:      int(sourceTx.ChainID),
			Status:       sourceTx.Status,
			NativeTxHash: sourceTx.TxHash,
			From:         sourceTx.FromAddr,
		},
	}, nil
}
