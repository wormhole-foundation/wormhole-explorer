package consumer

import (
	"context"
	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"strconv"
	"time"
)

type PostgreSQLRepository interface {
	UpsertOriginTx(ctx context.Context, params *UpsertOriginTxParams) error
	UpsertTargetTx(ctx context.Context, globalTx *TargetTxUpdate) error
	GetTxStatus(ctx context.Context, targetTxUpdate *TargetTxUpdate) (string, error)
	AlreadyProcessed(ctx context.Context, vaDigest string) (bool, error)
	RegisterProcessedVaa(ctx context.Context, vaaDigest, vaaId string) error
}

func NewPostgreSQLRepository(postreSQLClient *db.DB) *PostgreSQLUpsertTx {
	return &PostgreSQLUpsertTx{
		dbClient: postreSQLClient,
	}
}

type PostgreSQLUpsertTx struct {
	dbClient *db.DB
}

type noOpPostgreSQLUpsertTx struct{}

func (n *noOpPostgreSQLUpsertTx) AlreadyProcessed(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (n *noOpPostgreSQLUpsertTx) UpsertOriginTx(_ context.Context, _ *UpsertOriginTxParams) error {
	return nil
}

func (n *noOpPostgreSQLUpsertTx) UpsertTargetTx(_ context.Context, _ *TargetTxUpdate) error {
	return nil
}

func (n *noOpPostgreSQLUpsertTx) GetTxStatus(_ context.Context, _ *TargetTxUpdate) (string, error) {
	return "", nil
}

func (n *noOpPostgreSQLUpsertTx) RegisterProcessedVaa(ctx context.Context, _, _ string) error {
	return nil
}

func NoOpPostreSQLRepository() PostgreSQLRepository {
	return &noOpPostgreSQLUpsertTx{}
}

func (p *PostgreSQLUpsertTx) UpsertOriginTx(ctx context.Context, params *UpsertOriginTxParams) error {

	query := `
		INSERT INTO wormhole.wh_operation_transactions 
		(chain_id, tx_hash, type, created_at, updated_at, attestation_vaas_id, vaa_id, status, from_address, to_address, block_number, blockchain_method, fee, raw_fee, timestamp, rpc_response)  
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (chain_id, tx_hash) DO UPDATE
		SET 
			type = EXCLUDED.type,
			created_at = wormhole.wh_operation_transactions.created_at,
			updated_at = EXCLUDED.updated_at,
			attestation_vaas_id = EXCLUDED.attestation_vaas_id,
			vaa_id = EXCLUDED.vaa_id,
			status = EXCLUDED.status,
			from_address = COALESCE(EXCLUDED.from_address, wormhole.wh_operation_transactions.from_address),
			to_address = COALESCE(EXCLUDED.to_address, wormhole.wh_operation_transactions.to_address),
			block_number = COALESCE(EXCLUDED.block_number, wormhole.wh_operation_transactions.block_number),
			blockchain_method = COALESCE(EXCLUDED.blockchain_method, wormhole.wh_operation_transactions.blockchain_method),
			fee = COALESCE(EXCLUDED.fee, wormhole.wh_operation_transactions.fee),
			raw_fee = COALESCE(EXCLUDED.raw_fee, wormhole.wh_operation_transactions.raw_fee),
			timestamp = EXCLUDED.timestamp,
			rpc_response = COALESCE(EXCLUDED.rpc_response, wormhole.wh_operation_transactions.rpc_response)
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

func (p *PostgreSQLUpsertTx) UpsertTargetTx(ctx context.Context, params *TargetTxUpdate) error {
	query := `
		INSERT INTO wormhole.wh_operation_transactions 
		(chain_id, tx_hash, type, created_at, updated_at, attestation_vaas_id, vaa_id, status, from_address, to_address, block_number, blockchain_method, fee, raw_fee, timestamp, rpc_response)  
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (chain_id, tx_hash) DO UPDATE
		SET 
			type = EXCLUDED.type,
			created_at = wormhole.wh_operation_transactions.created_at,
			updated_at = EXCLUDED.updated_at,
			attestation_vaas_id = EXCLUDED.attestation_vaas_id,
			vaa_id = EXCLUDED.vaa_id,
			status = COALESCE(EXCLUDED.status, wormhole.wh_operation_transactions.status),
			from_address = COALESCE(EXCLUDED.from_address, wormhole.wh_operation_transactions.from_address),
			to_address = COALESCE(EXCLUDED.to_address, wormhole.wh_operation_transactions.to_address),
			block_number = COALESCE(EXCLUDED.block_number, wormhole.wh_operation_transactions.block_number),
			blockchain_method = COALESCE(EXCLUDED.blockchain_method, wormhole.wh_operation_transactions.blockchain_method),
			fee = COALESCE(EXCLUDED.fee, wormhole.wh_operation_transactions.fee),
			raw_fee = COALESCE(EXCLUDED.raw_fee, wormhole.wh_operation_transactions.raw_fee),
			timestamp = EXCLUDED.timestamp,
			rpc_response = COALESCE(EXCLUDED.rpc_response, wormhole.wh_operation_transactions.rpc_response)
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

func (p *PostgreSQLUpsertTx) GetTxStatus(ctx context.Context, targetTxUpdate *TargetTxUpdate) (string, error) {
	var status string
	err := p.dbClient.SelectOne(ctx, &status, `SELECT status FROM wormhole.wh_operation_transactions WHERE chain_id = $1 AND tx_hash = $2`, targetTxUpdate.Destination.ChainID, targetTxUpdate.Destination.TxHash)
	return status, err
}

func (p *PostgreSQLUpsertTx) AlreadyProcessed(ctx context.Context, vaDigest string) (bool, error) {
	var count int
	err := p.dbClient.SelectOne(ctx, &count, `SELECT COUNT(*) FROM wormhole.wh_operation_transactions_processed WHERE id = $1`, vaDigest)
	return count > 0, err
}

func (p *PostgreSQLUpsertTx) RegisterProcessedVaa(ctx context.Context, vaaDigest, vaaId string) error {
	now := time.Now()
	_, err := p.dbClient.Exec(ctx,
		`INSERT INTO wormhole.wh_operation_transactions_processed (id,vaa_id,processed,created_at,updated_at)
			VALUES ($1,$2,true,$3,$4)`, vaaDigest, vaaId, now, now)
	return err
}