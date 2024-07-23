package consumer

import (
	"context"
	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"time"
)

type PostgreSQLRepository interface {
	UpsertOriginTx(ctx context.Context, params *UpsertOriginTxParams) error
	UpsertTargetTx(ctx context.Context, globalTx *TargetTxUpdate) error
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

func (n *noOpPostgreSQLUpsertTx) UpsertOriginTx(ctx context.Context, params *UpsertOriginTxParams) error {
	return nil
}

func (n *noOpPostgreSQLUpsertTx) UpsertTargetTx(ctx context.Context, globalTx *TargetTxUpdate) error {
	return nil
}

func NoOpPostreSQLRepository(postreSQLClient *db.DB) PostgreSQLRepository {
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
	var from, to, blockNumber, rpcResponse, nativeTxHash *string
	if params.TxDetail != nil {
		from = &params.TxDetail.From
		to = &params.TxDetail.To
		blockNumber = &params.TxDetail.BlockNumber
		nativeTxHash = &params.TxDetail.NativeTxHash
		if params.TxDetail.FeeDetail != nil {
			fee = &params.TxDetail.FeeDetail.Fee
			rawFee = params.TxDetail.FeeDetail.RawFee
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
	var from, to, blockNumber, blockchainMethod, status, txHash *string
	var timestamp, updatedAt *time.Time
	var chainID *vaa.ChainID
	if params.Destination != nil {
		from = &params.Destination.From
		to = &params.Destination.To
		blockNumber = &params.Destination.BlockNumber
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
