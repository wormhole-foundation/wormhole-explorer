package storage

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/utils"
	"github.com/wormhole-foundation/wormhole-explorer/fly/event"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/metrics"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// PostgresRepository is a storage repository.
type PostgresRepository struct {
	db      *db.DB
	metrics metrics.Metrics
	// TODO: after migration move eventDispatcher to handlers.
	eventDispatcher event.EventDispatcher
	logger          *zap.Logger
}

// NewPostgresRepository creates a new storage repository.
func NewPostgresRepository(db *db.DB, metrics metrics.Metrics,
	eventDispatcher event.EventDispatcher, logger *zap.Logger) *PostgresRepository {
	return &PostgresRepository{
		db:              db,
		metrics:         metrics,
		eventDispatcher: eventDispatcher,
		logger:          logger,
	}
}

// UpsertObservation upserts an observation.
func (r *PostgresRepository) UpsertObservation(ctx context.Context, o *gossipv1.SignedObservation, saveTxHash bool) error {
	// id = {message_id}/{guardian_address}/{hash}
	id := fmt.Sprintf("%s/%s/%s", o.MessageId, hex.EncodeToString(o.Addr), hex.EncodeToString(o.Hash))

	// current time
	now := time.Now()

	// get emitterChainID, emitterAddress, sequence
	messageID := strings.Split(o.MessageId, "/")
	strEmitterChainID, emitterAddress, strSequence := messageID[0], messageID[1], messageID[2]
	chainIDUint64, err := strconv.ParseUint(strEmitterChainID, 10, 16)
	if err != nil {
		r.logger.Error("Error parsing chainId",
			zap.String("messageId", o.MessageId),
			zap.Error(err))

		return err
	}
	emitterChainID := sdk.ChainID(chainIDUint64)
	sequence, err := strconv.ParseUint(strSequence, 10, 64)
	if err != nil {
		r.logger.Error("Error parsing sequence",
			zap.String("messageId", o.MessageId),
			zap.Error(err))
		return err
	}

	// guardian address
	guardianAddress := utils.NormalizeHex(hex.EncodeToString(o.Addr))

	// hash
	hash := hex.EncodeToString(o.Hash)

	// native txHash
	txHash, err := domain.EncodeTrxHashByChainID(emitterChainID, o.GetTxHash())
	if err != nil {
		r.logger.Warn("Error encoding tx hash",
			zap.String("messageId", o.MessageId),
			zap.ByteString("txHash", o.GetTxHash()),
			zap.Error(err))

		r.metrics.IncObservationWithoutTxHash(emitterChainID)
	}

	query := `
		INSERT INTO wormhole.wh_observations 
		(id, emitter_chain_id, emitter_address, "sequence", hash, tx_hash, guardian_address, signature, created_at)  
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9) 
		ON CONFLICT(id) DO UPDATE 
		SET tx_hash = $6, signature = $8, updated_at = $10 
		RETURNING updated_at;
		`

	var result *time.Time
	err = r.db.ExecAndScan(ctx,
		&result,
		query,
		id,
		emitterChainID,
		emitterAddress,
		sequence,
		hash,
		txHash,
		guardianAddress,
		o.Signature,
		now,
		now)

	if err != nil {
		r.logger.Error("Error upserting observation",
			zap.String("messageId", o.MessageId),
			zap.Error(err))
		return err
	}

	rowInserted := isRowInserted(result)
	if rowInserted {
		r.metrics.IncObservationInserted(emitterChainID)
	}

	return nil
}

// UpsertVAA upserts a VAA.
func (r *PostgresRepository) UpsertVAA(ctx context.Context, v *sdk.VAA, serializedVaa []byte) error {
	id := utils.NormalizeHex(v.HexDigest()) //digest
	now := time.Now()

	table := "wormhole.wh_attestation_vaas"
	if v.EmitterChain == sdk.ChainIDPythNet {
		table = "wormhole.wh_attestation_vaas_pythnet"
	}

	queryTemplate := `
	INSERT INTO %s 
	(id, vaa_id, "version", emitter_chain_id, emitter_address, "sequence", guardian_set_index,
	raw, "timestamp", active, is_duplicated, created_at) 
	VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) 
	ON CONFLICT(id) DO UPDATE 
	SET vaa_id = $2, version =$3, emitter_chain_id = $4, emitter_address = $5, "sequence" = $6, guardian_set_index = $7, 
	raw = $8, "timestamp" = $9, updated_at = $13 
	RETURNING updated_at;
	`

	//RETURNING id, updated_at;
	query := fmt.Sprintf(queryTemplate, table)

	var result *time.Time
	err := r.db.ExecAndScan(ctx,
		&result,
		query,
		id,
		v.MessageID(),
		v.Version,
		v.EmitterChain,
		v.EmitterAddress,
		v.Sequence,
		v.GuardianSetIndex,
		serializedVaa,
		v.Timestamp,
		true,
		false,
		now,
		now)

	if err != nil {
		r.logger.Error("Error upserting VAA",
			zap.String("id", id),
			zap.String("vaaId", v.MessageID()),
			zap.Error(err))
		return err
	}

	rowInserted := isRowInserted(result)
	if rowInserted {
		r.metrics.IncVaaInserted(v.EmitterChain)

		vaa := event.Vaa{
			ID:               id,
			VaaID:            v.MessageID(),
			EmitterChainID:   uint16(v.EmitterChain),
			EmitterAddress:   v.EmitterAddress.String(),
			Sequence:         v.Sequence,
			Version:          v.Version,
			GuardianSetIndex: v.GuardianSetIndex,
			Raw:              serializedVaa,
			Timestamp:        v.Timestamp,
		}
		// dispatch new VAA event to the pipeline.
		// TODO:
		// -> define in spy component how to handle txHash because we dont have the txHash.
		// -> check mongo repo events.NewNotificationEvent[events.SignedVaa]
		err := r.eventDispatcher.NewVaa(ctx, vaa)
		if err != nil {
			r.logger.Error("Error dispatching new VAA event",
				zap.String("id", id),
				zap.String("vaaId", v.MessageID()),
				zap.Error(err))
			return err
		}
	}
	return nil
}

// UpsertHeartbeat upserts a heartbeat.
// Questions: sWe need to support this in the v2??
func (r *PostgresRepository) UpsertHeartbeat(hb *gossipv1.Heartbeat) error {
	id := utils.NormalizeHex(hb.GuardianAddr)
	now := time.Now()
	timestamp := time.Unix(0, hb.Timestamp)
	bootTimestamp := time.Unix(0, hb.BootTimestamp)

	query := `
	INSERT INTO wormhole.wh_heartbeats
	(id, guardian_name, boot_timestamp, "timestamp", version, networks, feature, created_at, updated_at)
	VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9) 
	ON CONFLICT(id) DO UPDATE 
	SET guardian_name = $2, boot_timestamp = $3, "timestamp" = $4, version = $5, networks = $6, feature = $7, updated_at = $9;
	`

	_, err := r.db.Exec(context.Background(),
		query,
		id,
		hb.GetNodeName(),
		bootTimestamp,
		timestamp,
		hb.Version,
		hb.Networks,
		hb.Features,
		now,
		now)

	if err != nil {
		r.logger.Error("Error upserting heartbeat",
			zap.String("id", id),
			zap.Error(err))
		return err
	}

	return nil
}

// UpsertGovernorConfig upserts a governor config.
func (r *PostgresRepository) UpsertGovernorConfig(ctx context.Context, govC *gossipv1.SignedChainGovernorConfig) error {
	// id is the guardian address.
	id := hex.EncodeToString(govC.GuardianAddr)
	now := time.Now()

	// unmarshal governor config
	var gc gossipv1.ChainGovernorConfig
	err := proto.Unmarshal(govC.Config, &gc)
	if err != nil {
		r.logger.Error("Error unmarshalling governor config",
			zap.String("id", id),
			zap.Error(err))
		return err
	}

	governorConfig := toGovernorConfigUpdate(&gc)
	timestamp := time.Unix(0, governorConfig.Timestamp)

	query := `
	INSERT INTO wormhole.wh_governor_config
	(id, guardian_name, counter, timestamp, tokens, created_at, updated_at)
	VALUES($1, $2, $3, $4, $5, $6, $7) 
	ON CONFLICT(id) DO UPDATE 
	SET counter = $3, timestamp = $4, tokens = $5, updated_at = $7;
	`

	_, err = r.db.Exec(context.Background(),
		query,
		id,
		gc.GetNodeName(),
		governorConfig.Counter,
		timestamp,
		governorConfig.Tokens,
		now,
		now)

	if err != nil {
		r.logger.Error("Error upserting governor config",
			zap.String("id", id),
			zap.Error(err))
		return err
	}

	// send governor config to topic. [fly-event-processor]
	errDispatcher := r.eventDispatcher.NewGovernorConfig(context.TODO(), event.GovernorConfig{
		NodeAddress: id,
		NodeName:    governorConfig.NodeName,
		Counter:     governorConfig.Counter,
		Timestamp:   governorConfig.Timestamp,
		Chains:      governorConfig.Chains,
	})

	if errDispatcher != nil {
		r.logger.Error("Error sending governor config to topic",
			zap.String("id", id),
			zap.Error(errDispatcher))
	}
	return errDispatcher
}

// UpsertGovernorStatus upserts a governor status.
func (r *PostgresRepository) UpsertGovernorStatus(ctx context.Context, govS *gossipv1.SignedChainGovernorStatus) error {
	// id is the guardian address.
	id := hex.EncodeToString(govS.GuardianAddr)
	now := time.Now()

	// unmarshal governor status
	var gs gossipv1.ChainGovernorStatus
	err := proto.Unmarshal(govS.Status, &gs)
	if err != nil {
		r.logger.Error("Error unmarshalling governor status",
			zap.String("id", id),
			zap.Error(err))
		return err
	}
	governorStatus := toGovernorStatusUpdate(&gs)
	timestamp := time.Unix(0, governorStatus.Timestamp)

	query := `
	INSERT INTO wormhole.wh_governor_status
	(id, guardian_name, message, timestamp, created_at, updated_at)
	VALUES($1, $2, $3, $4, $5, $6) 
	ON CONFLICT(id) DO UPDATE 
	SET message = $3, timestamp = $4, updated_at = $6;
	`

	_, err = r.db.Exec(context.Background(),
		query,
		id,
		gs.GetNodeName(),
		governorStatus,
		timestamp,
		now,
		now)

	if err != nil {
		r.logger.Error("Error upserting governor status",
			zap.String("id", id),
			zap.Error(err))
		return err
	}

	// send governor status to topic. [fly-event-processor]
	errDispatcher := r.eventDispatcher.NewGovernorStatus(context.TODO(), event.GovernorStatus{
		NodeAddress: id,
		NodeName:    governorStatus.NodeName,
		Counter:     governorStatus.Counter,
		Timestamp:   governorStatus.Timestamp,
		Chains:      governorStatus.Chains,
	})

	if errDispatcher != nil {
		r.logger.Error("Error sending governor status to topic",
			zap.String("id", id),
			zap.Error(errDispatcher))
	}
	return errDispatcher
}

// FindVaasByVaaID finds VAAs by VAA ID.
func (r *PostgresRepository) FindVaasByVaaID(ctx context.Context, vaaID string) ([]*AttestationVaa, error) {
	query := `
	SELECT id, vaa_id, "version", emitter_chain_id, emitter_address, "sequence", guardian_set_index,
	raw, "timestamp", active, is_duplicated, created_at, updated_at
	FROM wormhole.wh_attestation_vaas 
	WHERE vaa_id = $1;`

	var AttestationVaas []*AttestationVaa
	err := r.db.Select(ctx, &AttestationVaas, query, vaaID)
	if err != nil {
		r.logger.Error("Error finding vaas by vaaID",
			zap.String("vaaId", vaaID),
			zap.Error(err))
		return nil, err
	}

	return AttestationVaas, nil
}

// ReplaceVaaTxHash replaces a VAA transaction hash.
// Requiered method to support Storager interface
// TODO: delete this methods after migration
func (r *PostgresRepository) ReplaceVaaTxHash(ctx context.Context, vaaID string, oldTxHash string, newTxHash string) error {
	return nil
}

// FindVaaByID finds a VAA by ID.
// Requiered method to support Storager interface
// TODO: delete this methods after migration
func (r *PostgresRepository) FindVaaByID(ctx context.Context, vaaID string) (*VaaUpdate, error) {
	return nil, nil
}

// FindVaaByChainID finds a VAA by chain ID.
// Requiered method to support Storager interface
// TODO: delete this methods after migration
func (r *PostgresRepository) FindVaaByChainID(ctx context.Context, chainID sdk.ChainID, page int64, pageSize int64) ([]*VaaUpdate, error) {
	return nil, nil
}

// UpsertDuplicateVaa upserts a duplicate VAA.
// Requiered method to support Storager interface
// TODO: delete this methods after migration
func (r *PostgresRepository) UpsertDuplicateVaa(ctx context.Context, v *sdk.VAA, serializedVaa []byte) error {
	id := utils.NormalizeHex(v.HexDigest()) //digest
	now := time.Now()

	table := "wormhole.wh_attestation_vaas"
	if v.EmitterChain == sdk.ChainIDPythNet {
		table = "wormhole.wh_attestation_vaas_pythnet"
	}

	queryTemplate := `
	INSERT INTO %s 
	(id, vaa_id, "version", emitter_chain_id, emitter_address, "sequence", guardian_set_index,
	raw, "timestamp", active, is_duplicated, created_at) 
	VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) 
	ON CONFLICT(id) DO UPDATE 
	SET vaa_id = $2, version =$3, emitter_chain_id = $4, emitter_address = $5, "sequence" = $6, guardian_set_index = $7, 
	raw = $8, "timestamp" = $9, updated_at = $13 
	RETURNING updated_at;
	`

	//RETURNING id, updated_at;
	query := fmt.Sprintf(queryTemplate, table)

	var result *time.Time
	err := r.db.ExecAndScan(ctx,
		&result,
		query,
		id,
		v.MessageID(),
		v.Version,
		v.EmitterChain,
		v.EmitterAddress,
		v.Sequence,
		v.GuardianSetIndex,
		serializedVaa,
		v.Timestamp,
		false,
		true,
		now,
		now)

	if err != nil {
		r.logger.Error("Error upserting VAA",
			zap.String("id", id),
			zap.String("vaaId", v.MessageID()),
			zap.Error(err))
		return err
	}

	rowInserted := isRowInserted(result)
	if rowInserted {
		r.metrics.IncVaaInserted(v.EmitterChain)

		vaa := event.Vaa{
			ID:               id,
			VaaID:            v.MessageID(),
			EmitterChainID:   uint16(v.EmitterChain),
			EmitterAddress:   v.EmitterAddress.String(),
			Sequence:         v.Sequence,
			Version:          v.Version,
			GuardianSetIndex: v.GuardianSetIndex,
			Raw:              serializedVaa,
			Timestamp:        v.Timestamp,
		}
		// dispatch new VAA event to the pipeline.
		// TODO:
		// -> define in spy component how to handle txHash because we dont have the txHash.
		// -> check mongo repo events.NewNotificationEvent[events.SignedVaa]
		err := r.eventDispatcher.NewVaa(ctx, vaa)
		if err != nil {
			r.logger.Error("Error dispatching new VAA event",
				zap.String("id", id),
				zap.String("vaaId", v.MessageID()),
				zap.Error(err))
			return err
		}
	}
	return nil
}

// isRowInserted checks if a row was inserted.
func isRowInserted(result *time.Time) bool {
	isRowInserted := false
	if result == nil {
		isRowInserted = true
	}
	return isRowInserted
}
