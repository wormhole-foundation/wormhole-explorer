package consumer

import (
	"context"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/utils"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/chains"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/internal/repository/vaa"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type PostgreSQLRepository struct {
	dbClient      *db.DB
	vaaRepository *vaa.RepositoryPostreSQL
	metrics       metrics.Metrics
	logger        *zap.Logger
}

func NewPostgreSQLRepository(postreSQLClient *db.DB, vaaRepository *vaa.RepositoryPostreSQL,
	metrics metrics.Metrics, logger *zap.Logger) *PostgreSQLRepository {
	return &PostgreSQLRepository{
		metrics:       metrics,
		dbClient:      postreSQLClient,
		vaaRepository: vaaRepository,
		logger:        logger,
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

	p.metrics.IncOperationTxSourceInserted(uint16(originTx.ChainId))
	return nil
}

func (p *PostgreSQLRepository) upsertOriginTx(ctx context.Context, params *UpsertOriginTxParams) error {

	// upsert source tx.
	err := p.upsertSourceTx(ctx, params)
	if err != nil {
		return err
	}

	// upsert operation address.
	err = p.upsertOperationAddress(ctx, params)
	if err != nil {
		return err
	}

	// register processed vaa.
	err = p.registerProcessedVaa(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgreSQLRepository) upsertSourceTx(ctx context.Context, params *UpsertOriginTxParams) error {

	query := `
		INSERT INTO wormholescan.wh_operation_transactions 
		(chain_id, tx_hash, type, created_at, updated_at, attestation_id, message_id, status, from_address, to_address, block_number, blockchain_method, fee_detail, timestamp, rpc_response, source_event, track_id_event)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		ON CONFLICT (message_id, tx_hash) DO UPDATE
		SET 
			type = EXCLUDED.type,
			created_at = wormholescan.wh_operation_transactions.created_at,
			updated_at = EXCLUDED.updated_at,
			attestation_id = EXCLUDED.attestation_id,
			message_id = EXCLUDED.message_id,
			status = EXCLUDED.status,
			from_address = COALESCE(EXCLUDED.from_address, wormholescan.wh_operation_transactions.from_address),
			to_address = COALESCE(EXCLUDED.to_address, wormholescan.wh_operation_transactions.to_address),
			block_number = COALESCE(EXCLUDED.block_number, wormholescan.wh_operation_transactions.block_number),
			blockchain_method = COALESCE(EXCLUDED.blockchain_method, wormholescan.wh_operation_transactions.blockchain_method),
			fee_detail = COALESCE(EXCLUDED.fee_detail, wormholescan.wh_operation_transactions.fee_detail),
			timestamp = EXCLUDED.timestamp,
			rpc_response = COALESCE(EXCLUDED.rpc_response, wormholescan.wh_operation_transactions.rpc_response),
			source_event = COALESCE(EXCLUDED.source_event, wormholescan.wh_operation_transactions.source_event),
			track_id_event = COALESCE(EXCLUDED.track_id_event, wormholescan.wh_operation_transactions.track_id_event)
		`

	var from, to, rpcResponse, nativeTxHash *string
	var blockNumber string
	var feeDetail *chains.FeeDetail
	if params.TxDetail != nil {
		from = &params.TxDetail.From
		if params.TxDetail.NormalizedFrom != "" {
			from = &params.TxDetail.NormalizedFrom
		}
		to = &params.TxDetail.To
		if params.TxDetail.NormalizesTo != "" {
			to = &params.TxDetail.NormalizesTo
		}
		nativeTxHash = &params.TxDetail.NativeTxHash
		if params.TxDetail.NormalizedTxHash != "" {
			nativeTxHash = &params.TxDetail.NormalizedTxHash
		}

		if params.TxDetail.RpcResponse != "" {
			rpcResponse = &params.TxDetail.RpcResponse
		}

		if params.TxDetail.FeeDetail != nil {
			feeDetail = params.TxDetail.FeeDetail
		}
		blockNumber = params.TxDetail.BlockNumber
	}

	_, err := p.dbClient.Exec(ctx, query,
		params.ChainId,
		nativeTxHash,
		params.TxType,    // type
		time.Now(),       // created_at
		time.Now(),       // updated_at
		params.Id,        // attestation_id
		params.VaaId,     // message_id
		params.TxStatus,  // status
		from,             // from_address
		to,               // to_address
		blockNumber,      // block_number
		nil,              // blockchain_method: only applies for incoming targetTx from blockchain-watcher
		feeDetail,        // fee_detail
		params.Timestamp, // timestamp
		rpcResponse,      // rpc_response
		params.Source,    // source_event
		params.TrackID,   // track_id_event
	)

	if err != nil {
		return err
	}

	return nil
}

func (p *PostgreSQLRepository) upsertOperationAddress(ctx context.Context, params *UpsertOriginTxParams) error {
	// from address not found. can't upsert operation address.
	if params.TxDetail == nil {
		return nil
	}

	// from address is empty. can't upsert operation address.
	if params.TxDetail.From == "" {
		p.logger.Warn("from address is empty", zap.String("id", params.Id), zap.String("vaaId", params.VaaId))
		return nil
	}

	fromAddress := params.TxDetail.From
	if params.TxDetail.NormalizedFrom != "" {
		fromAddress = params.TxDetail.NormalizedFrom
	}

	// check timestamp can not be nil.
	if params.Timestamp == nil {
		p.logger.Error("timestamp is nil", zap.String("id", params.Id), zap.String("vaaId", params.VaaId))
		return nil
	}

	now := time.Now()
	_, err := p.dbClient.Exec(ctx,
		`INSERT INTO wormholescan.wh_operation_addresses (id, address, address_type, timestamp, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id, address) DO UPDATE SET address = $2, address_type = $3, timestamp = $4, updated_at = $6`,
		params.Id, fromAddress, "source", *params.Timestamp, now, now)

	return err
}

func (p *PostgreSQLRepository) registerProcessedVaa(ctx context.Context, params *UpsertOriginTxParams) error {
	// get normalized tx hash.
	var txHash string
	if params.TxDetail != nil {
		txHash = params.TxDetail.NativeTxHash
		if params.TxDetail.NormalizedTxHash != "" {
			txHash = params.TxDetail.NormalizedTxHash
		}
	}

	// tx hash is empty. can't register processed vaa.
	if txHash == "" {
		p.logger.Warn("tx hash is empty", zap.String("id", params.Id), zap.String("vaaId", params.VaaId))
		return nil
	}
	now := time.Now()

	// insert into wh_operation_transactions_processed.
	_, err := p.dbClient.Exec(ctx,
		`INSERT INTO wormholescan.wh_operation_transactions_processed (message_id,tx_hash,attestation_id,"type", processed,created_at,updated_at)
				VALUES ($1,$2,$3,$4,true,$5,$6)
				ON CONFLICT (message_id, tx_hash) DO UPDATE
					SET updated_at = EXCLUDED.updated_at`,
		params.VaaId, txHash, params.Id, "source", now, now)
	return err
}

func (p *PostgreSQLRepository) UpsertTargetTx(ctx context.Context, params *TargetTxUpdate) error {
	query := `
		INSERT INTO wormholescan.wh_operation_transactions 
		(chain_id, tx_hash, type, created_at, updated_at, attestation_id, message_id, status, from_address, to_address, block_number, blockchain_method, fee_detail, timestamp, rpc_response, source_event, track_id_event)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		ON CONFLICT (message_id, tx_hash) DO UPDATE
		SET 
			type = EXCLUDED.type,
			created_at = wormholescan.wh_operation_transactions.created_at,
			updated_at = EXCLUDED.updated_at,
			attestation_id = EXCLUDED.attestation_id,
			message_id = EXCLUDED.message_id,
			status = COALESCE(EXCLUDED.status, wormholescan.wh_operation_transactions.status),
			from_address = COALESCE(EXCLUDED.from_address, wormholescan.wh_operation_transactions.from_address),
			to_address = COALESCE(EXCLUDED.to_address, wormholescan.wh_operation_transactions.to_address),
			block_number = COALESCE(EXCLUDED.block_number, wormholescan.wh_operation_transactions.block_number),
			blockchain_method = COALESCE(EXCLUDED.blockchain_method, wormholescan.wh_operation_transactions.blockchain_method),
			fee_detail = COALESCE(EXCLUDED.fee_detail, wormholescan.wh_operation_transactions.fee_detail),
			timestamp = EXCLUDED.timestamp,
			rpc_response = COALESCE(EXCLUDED.rpc_response, wormholescan.wh_operation_transactions.rpc_response),
			source_event = COALESCE(EXCLUDED.source_event, wormholescan.wh_operation_transactions.source_event),
			track_id_event = COALESCE(EXCLUDED.track_id_event, wormholescan.wh_operation_transactions.track_id_event)
		`

	var from, to, blockchainMethod, status, txHash *string
	var blockNumber string
	var timestamp, updatedAt *time.Time
	var chainID *sdk.ChainID
	var feeDetail *FeeDetail
	if params.Destination != nil {
		from = &params.Destination.From
		if utils.StartsWith0x(params.Destination.From) {
			normalizedFrom := utils.Remove0x(params.Destination.From)
			from = &normalizedFrom
		}
		to = &params.Destination.To
		if utils.StartsWith0x(params.Destination.To) {
			normalizedTo := utils.Remove0x(params.Destination.To)
			to = &normalizedTo
		}
		blockchainMethod = &params.Destination.Method
		status = &params.Destination.Status
		txHash = &params.Destination.TxHash
		chainID = &params.Destination.ChainID
		timestamp = params.Destination.Timestamp
		updatedAt = params.Destination.UpdatedAt
		blockNumber = params.Destination.BlockNumber

		if params.Destination.FeeDetail != nil {
			feeDetail = params.Destination.FeeDetail
		}
	}
	_, err := p.dbClient.Exec(ctx, query,
		chainID,
		txHash,
		"target-tx",      // type
		time.Now(),       // created_at
		updatedAt,        // updated_at
		params.ID,        // attestation_id
		params.VaaID,     // message_id
		status,           // status
		from,             // from_address
		to,               // to_address
		blockNumber,      // block_number
		blockchainMethod, // blockchain_method
		feeDetail,        // fee_detail
		timestamp,        // timestamp
		nil,              // rpc_response
		params.Source,    // source_event
		params.TrackID,   // track_id_event

	)
	if err != nil {
		return err
	}

	if chainID != nil {
		p.metrics.IncOperationTxTargetInserted(uint16(*chainID))
	}
	return err
}

func (p *PostgreSQLRepository) GetTxStatus(ctx context.Context, targetTxUpdate *TargetTxUpdate) (string, error) {
	var status string
	err := p.dbClient.SelectOne(ctx, &status, `SELECT status FROM wormholescan.wh_operation_transactions WHERE chain_id = $1 AND tx_hash = $2`, targetTxUpdate.Destination.ChainID, targetTxUpdate.Destination.TxHash)
	return status, err
}

func (p *PostgreSQLRepository) AlreadyProcessed(ctx context.Context, vaaId string, txhash string) (bool, error) {
	var count int
	err := p.dbClient.SelectOne(ctx, &count, `SELECT COUNT(*) FROM wormholescan.wh_operation_transactions_processed WHERE message_id = $1 and tx_hash = $2`, vaaId, txhash)
	return count > 0, err
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
		ID       string `db:"message_id"`
		TxHash   string `db:"tx_hash"`
		ChainID  uint16 `db:"emitter_chain_id"`
		Status   string `db:"status"`
		FromAddr string `db:"from_address"`
	}

	query := `
	SELECT o.message_id, .o.tx_hash, v.emitter_chain_id, o.status, o.from_address
	FROM wormholescan.wh_attestation_vaas as v
	INNER JOIN wormholescan.wh_operation_transactions as o ON o.id = v.id 
	WHERE v.message_id = $1 and v.active = true and o.type = 'source-tx'
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

// GetIDByVaaID returns the id for the given vaa id
func (p *PostgreSQLRepository) GetIDByVaaID(ctx context.Context, vaaID string) (string, error) {

	v, err := p.vaaRepository.GetVaa(ctx, vaaID)
	if err != nil {
		return "", err
	}

	return domain.GetDigestFromRaw(v.Vaa)
}
