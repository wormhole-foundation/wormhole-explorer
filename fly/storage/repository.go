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
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// Repository is a storage repository.
type Repository struct {
	db     *db.DB
	logger *zap.Logger
}

// // NewRepository creates a new storage repository.
func NewRepository(db *db.DB, logger *zap.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

// UpsertObservation upserts an observation.
func (r *Repository) UpsertObservation(ctx context.Context, o *gossipv1.SignedObservation, saveTxHash bool) error {
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

	// Filter pytnet observations
	// TODO: check if we can filter pyth observations before push observation to internal queue.
	if emitterChainID == sdk.ChainIDPythNet {
		return nil
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

		//TODO: review metrics.IncObservationWithoutTxHash(chainID)
	}

	// TODO: add upsert logic + check schema wormhole.
	query := `
		INSERT INTO wormhole.wh_observations 
		(id, emitter_chain_id, emitter_address, "sequence", hash, tx_hash, guardian_address, signature, created_at, updated_at)  
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
		`
	_, err = r.db.Exec(ctx,
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

	//TODO metrics.IncObservationInserted(emitterChainID)
	return nil
}

// UpsertVAA upserts a VAA.
func (r *Repository) UpsertVAA(ctx context.Context, v *sdk.VAA, serializedVaa []byte) error {
	id := utils.NormalizeHex(v.HexDigest()) //digest
	now := time.Now()

	query := `
	INSERT INTO wormhole.wh_attestation_vaas 
	(id, vaa_id, "version", emitter_chain_id, emitter_address, "sequence", guardian_set_index,
	raw, "timestamp", active, is_duplicated, created_at, updated_at) 
	VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);
	`
	// out of
	_, err := r.db.Exec(ctx,
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
		true,  // TODO: define if we handle this field here o in fly-event-processor
		false, // TODO: define how handle this field.
		now,
		now)

	if err != nil {
		r.logger.Error("Error upserting VAA",
			zap.String("id", id),
			zap.String("vaaId", v.MessageID()),
			zap.Error(err))
		return err
	}

	//TODO metrics.IncVAAInserted(v.EmitterChain)
	return nil
}

// UpsertHeartbeat upserts a heartbeat.
// Questions: sWe need to support this in the v2??
func (r *Repository) UpsertHeartbeat(hb *gossipv1.Heartbeat) error {
	return nil
}

// UpsertGovernorConfig upserts a governor config.
func (r *Repository) UpsertGovernorConfig(ctx context.Context, govC *gossipv1.SignedChainGovernorConfig) error {
	return nil
}

// UpsertGovernorStatus upserts a governor status.
func (r *Repository) UpsertGovernorStatus(ctx context.Context, govS *gossipv1.SignedChainGovernorStatus) error {
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

	query := `
	INSERT INTO wormhole.wh_governor_status
	(id, guardian_name, message, created_at, updated_at)
	VALUES($1, $2, $3, $4, $5);
	`
	// TODO: upsert logic

	_, err = r.db.Exec(context.Background(),
		query,
		id,
		gs.GetNodeName(),
		governorStatus,
		now,
		now)

	if err != nil {
		r.logger.Error("Error upserting governor status",
			zap.String("id", id),
			zap.Error(err))
		return err
	}

	return nil
}

// ReplaceVaaTxHash replaces a VAA transaction hash.
// Requiered method to support Storager interface
// TODO: delete this methods after migration
func (r *Repository) ReplaceVaaTxHash(ctx context.Context, vaaID string, oldTxHash string, newTxHash string) error {
	return nil
}

// FindVaaByID finds a VAA by ID.
// Requiered method to support Storager interface
// TODO: delete this methods after migration
func (r *Repository) FindVaaByID(ctx context.Context, vaaID string) (*VaaUpdate, error) {
	return nil, nil
}

// FindVaaByChainID finds a VAA by chain ID.
// Requiered method to support Storager interface
// TODO: delete this methods after migration
func (r *Repository) FindVaaByChainID(ctx context.Context, chainID sdk.ChainID, page int64, pageSize int64) ([]*VaaUpdate, error) {
	return nil, nil
}

// UpsertDuplicateVaa upserts a duplicate VAA.
// Requiered method to support Storager interface
// TODO: delete this methods after migration
func (r *Repository) UpsertDuplicateVaa(ctx context.Context, v *sdk.VAA, serializedVaa []byte) error {
	return nil
}
