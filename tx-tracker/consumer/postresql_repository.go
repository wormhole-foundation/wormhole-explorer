package consumer

import (
	"context"
	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"time"
)

type PostgreSQLRepository interface {
	UpsertOriginTx(ctx context.Context, params *UpsertOriginTxParams) error
	UpsertTargetTx(ctx context.Context, globalTx *TargetTxUpdate) error
}

func NewPostgreSQLRepository(ctx context.Context, databaseURL string) (PostgreSQLRepository, error) {

	postreSQLClient, err := db.NewDB(ctx, databaseURL)
	if err != nil {
		return nil, err
	}

	return &postreSQLRepository{
		dbClient: postreSQLClient,
	}, err
}

type postreSQLRepository struct {
	dbClient *db.DB
}

func (p *postreSQLRepository) UpsertOriginTx(ctx context.Context, params *UpsertOriginTxParams) error {

	query := `
		INSERT INTO wormhole.wh_operation_transactions 
		(chain_id, tx_hash, type, created_at, updated_at, attestation_vaas_id, vaa_id, status, from_address, to_address, block_number, blockchain_method, fee, raw_fee, timestamp, rpc_response)  
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (chain_id, tx_hash) DO UPDATE
		SET 
			type = EXCLUDED.type,
			created_at = EXCLUDED.created_at,
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
	_, err := p.dbClient.Exec(ctx, query,
		params.ChainId,
		params.TxDetail.NativeTxHash,
		"source-tx",                         // type
		params.Timestamp,                    // created_at
		time.Now(),                          // updated_at
		params.VaaId,                        // attestation_vaas_id
		params.VaaId,                        // vaa_id
		params.TxStatus,                     // status
		params.TxDetail.From,                // from_address
		params.TxDetail.To,                  // to_address
		params.TxDetail.BlockNumber,         // block_number : todo: convert string to decimal(20,0)
		params.TxDetail.BlockchainRPCMethod, // blockchain_method
		params.TxDetail.FeeDetail.Fee,       // fee
		params.TxDetail.FeeDetail.RawFee,    // raw_fee : todo: CHECK IF IT REQUIRES MARSHALLING BEFORE OR NOT.
		params.Timestamp,                    // timestamp
		params.TxDetail.RpcResponse,         // rpc_response
	)

	return err
}

func (p *postreSQLRepository) UpsertTargetTx(ctx context.Context, params *TargetTxUpdate) error {
	query := `
		INSERT INTO wormhole.wh_operation_transactions 
		(chain_id, tx_hash, type, created_at, updated_at, attestation_vaas_id, vaa_id, status, from_address, to_address, block_number, blockchain_method, fee, raw_fee, timestamp, rpc_response)  
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (chain_id, tx_hash) DO UPDATE
		SET 
			type = EXCLUDED.type,
			created_at = EXCLUDED.created_at,
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
	_, err := p.dbClient.Exec(ctx, query,
		params.Destination.ChainID,
		params.Destination.TxHash,
		"target-tx",                         // type
		params.Destination.Timestamp,        // created_at
		params.Destination.UpdatedAt,        // updated_at
		params.ID,                           // attestation_vaas_id
		params.ID,                           // vaa_id
		params.Destination.Status,           // status
		params.Destination.From,             // from_address
		params.Destination.To,               // to_address
		params.Destination.BlockNumber,      // block_number : todo: convert string to decimal(20,0)
		params.Destination.Method,           // blockchain_method
		params.Destination.FeeDetail.Fee,    // fee
		params.Destination.FeeDetail.RawFee, // raw_fee : todo: CHECK IF IT REQUIRES MARSHALLING BEFORE OR NOT.
		params.Destination.Timestamp,        // timestamp
		nil,                                 // rpc_response
	)

	return err
}
