package storage

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
	eth_common "github.com/ethereum/go-ethereum/common"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/events"
	"github.com/wormhole-foundation/wormhole-explorer/common/repository"
	"github.com/wormhole-foundation/wormhole-explorer/common/utils"
	"github.com/wormhole-foundation/wormhole-explorer/fly/event"
	flyAlert "github.com/wormhole-foundation/wormhole-explorer/fly/internal/alert"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/track"
	"github.com/wormhole-foundation/wormhole-explorer/fly/producer"
	"github.com/wormhole-foundation/wormhole-explorer/fly/txhash"

	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// TODO remove repository after switch to postgres db.
// TODO separate and maybe share between fly and web
type MongoRepository struct {
	alertClient     alert.AlertClient
	metrics         metrics.Metrics
	db              *mongo.Database
	afterUpdate     producer.PushFunc
	txHashStore     txhash.TxHashStore
	eventDispatcher event.EventDispatcher
	log             *zap.Logger
	collections     struct {
		vaas           *mongo.Collection
		heartbeats     *mongo.Collection
		observations   *mongo.Collection
		governorConfig *mongo.Collection
		governorStatus *mongo.Collection
		vaasPythnet    *mongo.Collection
		vaaCounts      *mongo.Collection
		duplicateVaas  *mongo.Collection
	}
}

// TODO wrap repository with a service that filters using redis
func NewMongoRepository(alertService alert.AlertClient, metrics metrics.Metrics,
	db *mongo.Database,
	vaaTopicFunc producer.PushFunc,
	txHashStore txhash.TxHashStore,
	eventDispatcher event.EventDispatcher,
	log *zap.Logger) *MongoRepository {
	return &MongoRepository{alertService, metrics, db, vaaTopicFunc, txHashStore, eventDispatcher, log, struct {
		vaas           *mongo.Collection
		heartbeats     *mongo.Collection
		observations   *mongo.Collection
		governorConfig *mongo.Collection
		governorStatus *mongo.Collection
		vaasPythnet    *mongo.Collection
		vaaCounts      *mongo.Collection
		duplicateVaas  *mongo.Collection
	}{
		vaas:           db.Collection(repository.Vaas),
		heartbeats:     db.Collection("heartbeats"),
		observations:   db.Collection(repository.Observations),
		governorConfig: db.Collection("governorConfig"),
		governorStatus: db.Collection("governorStatus"),
		vaasPythnet:    db.Collection("vaasPythnet"),
		vaaCounts:      db.Collection("vaaCounts"),
		duplicateVaas:  db.Collection(repository.DuplicateVaas)}}
}

func (s *MongoRepository) UpsertVAA(ctx context.Context, v *sdk.VAA, serializedVaa []byte, active bool, isDuplicated bool) error {
	id := v.MessageID()
	now := time.Now()
	vaaDoc := &VaaUpdate{
		ID:               v.MessageID(),
		Timestamp:        &v.Timestamp,
		Version:          v.Version,
		EmitterChain:     v.EmitterChain,
		EmitterAddr:      v.EmitterAddress.String(),
		Sequence:         strconv.FormatUint(v.Sequence, 10),
		GuardianSetIndex: v.GuardianSetIndex,
		Vaa:              serializedVaa,
		Digest:           utils.NormalizeHex(v.HexDigest()),
		UpdatedAt:        &now,
	}

	update := bson.M{
		"$set":         vaaDoc,
		"$setOnInsert": indexedAt(now),
		"$inc":         bson.D{{Key: "revision", Value: 1}},
	}

	opts := options.Update().SetUpsert(true)
	var err error
	var result *mongo.UpdateResult
	if sdk.ChainIDPythNet == v.EmitterChain {
		result, err = s.collections.vaasPythnet.UpdateByID(ctx, id, update, opts)
		if err != nil {
			// send alert when exists an error saving ptth vaa.
			alertContext := alert.AlertContext{
				Details: vaaDoc.ToMap(),
				Error:   err,
			}
			s.alertClient.CreateAndSend(ctx, flyAlert.ErrorSavePyth, alertContext)
		}
	} else {
		uniqueVaaID := domain.CreateUniqueVaaID(v)
		txHash, err := s.txHashStore.Get(ctx, uniqueVaaID)
		if err != nil {
			s.log.Warn("Finding vaaIdTxHash", zap.String("id", id), zap.Error(err))
		}
		if txHash != nil {
			vaaDoc.TxHash = *txHash
		}
		result, err = s.collections.vaas.UpdateByID(ctx, id, update, opts)
		if err != nil {
			// send alert when exists an error saving vaa.
			alertContext := alert.AlertContext{
				Details: vaaDoc.ToMap(),
				Error:   err,
			}
			s.alertClient.CreateAndSend(ctx, flyAlert.ErrorSaveVAA, alertContext)
		}
	}
	if err == nil && s.isNewRecord(result) {
		s.metrics.IncVaaInserted(v.EmitterChain)
		s.updateVAACount(v.EmitterChain)

		// send signedvaa event to topic.
		event, newErr := events.NewNotificationEvent[events.SignedVaa](
			track.GetTrackID(v.MessageID()), "fly", events.SignedVaaType,
			events.SignedVaa{
				ID:               v.MessageID(),
				EmitterChain:     uint16(v.EmitterChain),
				EmitterAddress:   v.EmitterAddress.String(),
				Sequence:         v.Sequence,
				GuardianSetIndex: v.GuardianSetIndex,
				Timestamp:        v.Timestamp,
				Vaa:              serializedVaa,
				TxHash:           vaaDoc.TxHash,
				Version:          int(v.Version),
			})
		if newErr != nil {
			return newErr
		}
		err = s.afterUpdate(ctx, &producer.Notification{ID: v.MessageID(), Event: event, EmitterChain: v.EmitterChain})
	}
	return err
}

func (s *MongoRepository) UpsertObservation(ctx context.Context, o *gossipv1.SignedObservation, saveTxHash bool) error {
	vaaID := strings.Split(o.MessageId, "/")
	chainIDStr, emitter, sequenceStr := vaaID[0], vaaID[1], vaaID[2]
	id := fmt.Sprintf("%s/%s/%s", o.MessageId, hex.EncodeToString(o.Addr), hex.EncodeToString(o.Hash))
	now := time.Now()

	chainIDUint, err := strconv.ParseUint(chainIDStr, 10, 16)
	if err != nil {
		s.log.Error("Error parsing chainId", zap.Error(err))
		return err
	}

	// TODO should we notify the caller that pyth observations are not stored?
	if sdk.ChainID(chainIDUint) == sdk.ChainIDPythNet {
		return nil
	}
	sequence, err := strconv.ParseUint(sequenceStr, 10, 64)
	if err != nil {
		s.log.Error("Error parsing sequence", zap.Error(err))
		return err
	}

	chainID := sdk.ChainID(chainIDUint)
	var nativeTxHash string
	switch chainID {
	case sdk.ChainIDSolana,
		sdk.ChainIDWormchain,
		sdk.ChainIDAptos:
	default:
		nativeTxHash, _ = domain.EncodeTrxHashByChainID(chainID, o.GetTxHash())
	}

	addr := eth_common.BytesToAddress(o.GetAddr())
	obs := ObservationUpdate{
		ChainID:      chainID,
		Emitter:      emitter,
		Sequence:     strconv.FormatUint(sequence, 10),
		MessageID:    o.GetMessageId(),
		Hash:         o.GetHash(),
		TxHash:       o.GetTxHash(),
		NativeTxHash: nativeTxHash,
		GuardianAddr: addr.String(),
		Signature:    o.GetSignature(),
		UpdatedAt:    &now,
	}

	update := bson.M{
		"$set":         obs,
		"$setOnInsert": indexedAt(now),
		"$inc":         bson.D{{Key: "revision", Value: 1}},
	}
	opts := options.Update().SetUpsert(true)
	_, err = s.collections.observations.UpdateByID(ctx, id, update, opts)
	if err != nil {
		s.log.Error("Error inserting observation", zap.Error(err))
		// send alert when exists an error saving observation.
		alertContext := alert.AlertContext{
			Details: obs.ToMap(),
			Error:   err,
		}
		s.alertClient.CreateAndSend(ctx, flyAlert.ErrorSaveObservation, alertContext)
		return err
	}

	s.metrics.IncObservationInserted(sdk.ChainID(chainID))

	if saveTxHash {

		txHash, err := domain.EncodeTrxHashByChainID(sdk.ChainID(chainID), o.GetTxHash())
		if err != nil {
			s.log.Warn("Error encoding tx hash",
				zap.Uint64("chainId", chainIDUint),
				zap.ByteString("txHash", o.GetTxHash()),
				zap.Error(err))
			s.metrics.IncObservationWithoutTxHash(chainID)
		}

		vaaTxHash := txhash.TxHash{
			ChainID:  chainID,
			Emitter:  emitter,
			Sequence: strconv.FormatUint(sequence, 10),
			TxHash:   txHash,
		}

		uniqueVaaID := domain.CreateUniqueVaaIDByObservation(o)
		err = s.txHashStore.Set(ctx, uniqueVaaID, vaaTxHash)
		if err != nil {
			s.log.Error("Error setting txHash", zap.Error(err))
			return err
		}
	}

	return err
}

func (s *MongoRepository) ReplaceVaaTxHash(ctx context.Context, vaaID, oldTxHash, newTxHash string) error {
	now := time.Now()
	update := bson.D{
		{Key: "$set", Value: bson.D{{Key: "txHash", Value: newTxHash}}},
		{Key: "$set", Value: bson.D{{Key: "_originTxHash", Value: oldTxHash}}},
		{Key: "$set", Value: bson.D{{Key: "updatedAt", Value: now}}},
	}
	_, err := s.collections.vaas.UpdateByID(ctx, vaaID, update)
	if err != nil {
		return nil
	}
	return nil
}

func (s *MongoRepository) UpsertHeartbeat(hb *gossipv1.Heartbeat) error {
	id := hb.GuardianAddr
	now := time.Now()
	update := bson.D{{Key: "$set", Value: hb}, {Key: "$set", Value: bson.D{{Key: "updatedAt", Value: now}}}, {Key: "$setOnInsert", Value: bson.D{{Key: "indexedAt", Value: now}}}}
	opts := options.Update().SetUpsert(true)
	_, err := s.collections.heartbeats.UpdateByID(context.TODO(), id, update, opts)
	if err != nil {
		s.log.Error("Error inserting heartbeat", zap.Error(err))
		// send alert when exists an error saving heartbeat.
		alertContext := alert.AlertContext{
			Details: map[string]string{
				"guardianAddr": hb.GuardianAddr,
				"nodeName":     hb.NodeName,
			},
			Error: err,
		}
		s.alertClient.CreateAndSend(context.TODO(), flyAlert.ErrorSaveHeartbeat, alertContext)
	}
	return err
}

func (s *MongoRepository) UpsertGovernorConfig(ctx context.Context, govC *gossipv1.SignedChainGovernorConfig) error {
	id := hex.EncodeToString(govC.GuardianAddr)
	now := time.Now()
	var gCfg gossipv1.ChainGovernorConfig
	err := proto.Unmarshal(govC.Config, &gCfg)
	if err != nil {
		s.log.Error("Error unmarshalling govr config", zap.Error(err))
		return err
	}

	cfg := toGovernorConfigUpdate(&gCfg)

	update := bson.D{{Key: "$set", Value: govC}, {Key: "$set", Value: bson.D{{Key: "parsedConfig", Value: cfg}}}, {Key: "$set", Value: bson.D{{Key: "updatedAt", Value: now}}}, {Key: "$setOnInsert", Value: bson.D{{Key: "createdAt", Value: now}}}}

	opts := options.Update().SetUpsert(true)
	_, err2 := s.collections.governorConfig.UpdateByID(context.TODO(), id, update, opts)

	if err2 != nil {
		s.log.Error("Error inserting govr cfg", zap.Error(err2))
		// send alert when exists an error saving governor config.
		alertContext := alert.AlertContext{
			Details: map[string]string{
				"nodeName": cfg.NodeName,
			},
			Error: err2,
		}
		s.alertClient.CreateAndSend(context.TODO(), flyAlert.ErrorSaveGovernorConfig, alertContext)
	}
	return err2
}

func (s *MongoRepository) UpsertGovernorStatus(ctx context.Context, govS *gossipv1.SignedChainGovernorStatus) error {
	id := hex.EncodeToString(govS.GuardianAddr)
	now := time.Now()
	var gStatus gossipv1.ChainGovernorStatus
	err := proto.Unmarshal(govS.Status, &gStatus)
	if err != nil {
		s.log.Error("Error unmarshalling govr status", zap.Error(err))
		return err
	}

	status := toGovernorStatusUpdate(&gStatus)

	update := bson.D{{Key: "$set", Value: govS}, {Key: "$set", Value: bson.D{{Key: "parsedStatus", Value: status}}}, {Key: "$set", Value: bson.D{{Key: "updatedAt", Value: now}}}, {Key: "$setOnInsert", Value: bson.D{{Key: "createdAt", Value: now}}}}

	opts := options.Update().SetUpsert(true)
	_, err2 := s.collections.governorStatus.UpdateByID(context.TODO(), id, update, opts)

	if err2 != nil {
		s.log.Error("Error inserting govr status", zap.Error(err2))
		// send alert when exists an error saving governor status.
		alertContext := alert.AlertContext{
			Details: map[string]string{
				"nodeName": status.NodeName,
			},
			Error: err2,
		}
		s.alertClient.CreateAndSend(context.TODO(), flyAlert.ErrorSaveGovernorStatus, alertContext)
		return err2
	}

	// send governor status to topic.
	err3 := s.eventDispatcher.NewGovernorStatus(context.TODO(), event.GovernorStatus{
		NodeAddress: id,
		NodeName:    status.NodeName,
		Counter:     status.Counter,
		Timestamp:   status.Timestamp,
		Chains:      status.Chains,
	})

	if err3 != nil {
		s.log.Error("Error sending governor status to topic",
			zap.String("guardian", status.NodeName),
			zap.Error(err3))
	}
	return err3
}

func (s *MongoRepository) updateVAACount(chainID sdk.ChainID) {
	update := bson.D{{Key: "$inc", Value: bson.D{{Key: "count", Value: uint64(1)}}}}
	opts := options.Update().SetUpsert(true)
	_, _ = s.collections.vaaCounts.UpdateByID(context.TODO(), chainID, update, opts)
}

func (s *MongoRepository) isNewRecord(result *mongo.UpdateResult) bool {
	return result.MatchedCount == 0 && result.ModifiedCount == 0 && result.UpsertedCount == 1
}

// GetMongoStatus get mongo server status
func (r *MongoRepository) GetMongoStatus(ctx context.Context) (*MongoStatus, error) {
	command := bson.D{{Key: "serverStatus", Value: 1}}
	result := r.db.RunCommand(ctx, command)
	if result.Err() != nil {
		return nil, result.Err()
	}

	var mongoStatus MongoStatus
	err := result.Decode(&mongoStatus)
	if err != nil {
		return nil, err
	}
	return &mongoStatus, nil
}

func toGovernorStatusUpdate(s *gossipv1.ChainGovernorStatus) *GovernorStatusUpdate {
	var chains []*ChainGovernorStatusChain
	for _, c := range s.Chains {
		var emitters []*ChainGovernorStatusEmitter
		for _, e := range c.Emitters {
			var enqueuedVaas []*ChainGovernorStatusEnqueuedVAA
			for _, ev := range e.EnqueuedVaas {
				enqueuedVaa := &ChainGovernorStatusEnqueuedVAA{
					Sequence:      strconv.FormatUint(ev.Sequence, 10),
					ReleaseTime:   ev.ReleaseTime,
					NotionalValue: Uint64(ev.NotionalValue),
					TxHash:        ev.TxHash,
				}
				enqueuedVaas = append(enqueuedVaas, enqueuedVaa)
			}

			emitter := &ChainGovernorStatusEmitter{
				EmitterAddress:    e.EmitterAddress,
				TotalEnqueuedVaas: Uint64(e.TotalEnqueuedVaas),
				EnqueuedVaas:      enqueuedVaas,
			}
			emitters = append(emitters, emitter)
		}

		chain := &ChainGovernorStatusChain{
			ChainId:                    c.ChainId,
			RemainingAvailableNotional: Uint64(c.RemainingAvailableNotional),
			Emitters:                   emitters,
		}
		chains = append(chains, chain)
	}

	status := &GovernorStatusUpdate{
		NodeName:  s.NodeName,
		Counter:   s.Counter,
		Timestamp: s.Timestamp,
		Chains:    chains,
	}
	return status
}

func toGovernorConfigUpdate(c *gossipv1.ChainGovernorConfig) *ChainGovernorConfigUpdate {

	var chains []*ChainGovernorConfigChain
	for _, c := range c.Chains {
		chain := &ChainGovernorConfigChain{
			ChainId:            c.ChainId,
			NotionalLimit:      Uint64(c.NotionalLimit),
			BigTransactionSize: Uint64(c.BigTransactionSize),
		}
		chains = append(chains, chain)
	}

	var tokens []*ChainGovernorConfigToken
	for _, t := range c.Tokens {
		token := &ChainGovernorConfigToken{
			OriginChainId: t.OriginChainId,
			OriginAddress: t.OriginAddress,
			Price:         t.Price,
		}
		tokens = append(tokens, token)
	}

	return &ChainGovernorConfigUpdate{
		NodeName:  c.NodeName,
		Counter:   c.Counter,
		Timestamp: c.Timestamp,
		Chains:    chains,
		Tokens:    tokens,
	}
}

func (r *MongoRepository) FindVaaByChainID(ctx context.Context, chainID sdk.ChainID, page int64, pageSize int64) ([]*VaaUpdate, error) {
	filter := bson.M{
		"emitterChain": chainID,
	}
	skip := page * pageSize
	opts := &options.FindOptions{Skip: &skip, Limit: &pageSize, Sort: bson.M{"timestamp": 1}}
	cur, err := r.collections.vaas.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	var result []*VaaUpdate
	err = cur.All(ctx, &result)
	return result, err
}

func (r *MongoRepository) FindVaaByID(ctx context.Context, vaaID string) (*VaaUpdate, error) {
	var vaa VaaUpdate
	if err := r.collections.vaas.FindOne(ctx, bson.M{"_id": vaaID}).Decode(&vaa); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &vaa, nil
}

func (s *MongoRepository) UpsertDuplicateVaa(ctx context.Context, v *sdk.VAA, serializedVaa []byte) error {
	if sdk.ChainIDPythNet == v.EmitterChain {
		return nil
	}

	uniqueVaaID := domain.CreateUniqueVaaID(v)
	now := time.Now()

	duplicateVaaDoc := createDuplicateVaaUpdateFromVaa(uniqueVaaID, v, serializedVaa, now)
	update := bson.M{
		"$set":         duplicateVaaDoc,
		"$setOnInsert": indexedAt(now),
		"$inc":         bson.D{{Key: "revision", Value: 1}},
	}

	opts := options.Update().SetUpsert(true)

	// TODO find by vaaID+vaaHash??
	txHash, err := s.txHashStore.Get(ctx, uniqueVaaID)
	if err != nil {
		s.log.Warn("Finding vaaIdTxHash", zap.String("id", uniqueVaaID), zap.Error(err))
	}
	if txHash != nil {
		duplicateVaaDoc.TxHash = *txHash
	}

	// Save duplicate vaa in duplicateVaas collection
	result, err := s.collections.duplicateVaas.UpdateByID(ctx, uniqueVaaID, update, opts)
	if err != nil {
		alertContext := alert.AlertContext{
			Details: duplicateVaaDoc.ToMap(),
			Error:   err,
		}
		s.alertClient.CreateAndSend(ctx, flyAlert.ErrorSaveDuplicateVAA, alertContext)
		return err
	}

	// Update isDuplicated field in vaas collection
	updateIsDuplicated := bson.D{
		{Key: "$set", Value: bson.D{{Key: "isDuplicated", Value: true}}},
		{Key: "$set", Value: bson.D{{Key: "updatedAt", Value: now}}},
	}
	_, err = s.collections.vaas.UpdateByID(ctx, v.MessageID(), updateIsDuplicated)
	if err != nil {
		alertContext := alert.AlertContext{
			Details: duplicateVaaDoc.ToMap(),
			Error:   err,
		}
		s.alertClient.CreateAndSend(ctx, flyAlert.ErrorSaveDuplicateVAA, alertContext)
		return err
	}

	// send signedvaa event to topic.
	if s.isNewRecord(result) {
		err := s.notifyNewVaa(ctx, v, serializedVaa, duplicateVaaDoc.TxHash)
		if err != nil {
			return err
		}
		return s.eventDispatcher.NewDuplicateVaa(ctx, event.DuplicateVaa{
			VaaID:            v.MessageID(),
			ChainID:          uint16(v.EmitterChain),
			Version:          v.Version,
			GuardianSetIndex: v.GuardianSetIndex,
			Vaa:              serializedVaa,
			Digest:           utils.NormalizeHex(v.HexDigest()),
			ConsistencyLevel: v.ConsistencyLevel,
			Timestamp:        &v.Timestamp,
		})
	}

	return nil
}

// FindDuplicateVaaByVaaID find vaas by vaaID
// Requiered method to support Storager interface
func (s *MongoRepository) FindVaasByVaaID(ctx context.Context, vaaID string) ([]*AttestationVaa, error) {
	// not implemented
	return nil, nil
}

func (s *MongoRepository) notifyNewVaa(ctx context.Context, v *sdk.VAA, serializedVaa []byte, txHash string) error {
	s.metrics.IncVaaInserted(v.EmitterChain)
	s.updateVAACount(v.EmitterChain)
	event, newErr := events.NewNotificationEvent[events.SignedVaa](
		track.GetTrackID(v.MessageID()), "fly", events.SignedVaaType,
		events.SignedVaa{
			ID:               v.MessageID(),
			EmitterChain:     uint16(v.EmitterChain),
			EmitterAddress:   v.EmitterAddress.String(),
			Sequence:         v.Sequence,
			GuardianSetIndex: v.GuardianSetIndex,
			Timestamp:        v.Timestamp,
			Vaa:              serializedVaa,
			TxHash:           txHash,
			Version:          int(v.Version),
		})
	if newErr != nil {
		return newErr
	}
	return s.afterUpdate(ctx, &producer.Notification{ID: v.MessageID(), Event: event, EmitterChain: v.EmitterChain})
}

func createDuplicateVaaUpdateFromVaa(uniqueID string, v *sdk.VAA, serializedVaa []byte, t time.Time) *DuplicateVaaUpdate {
	return &DuplicateVaaUpdate{
		ID:               uniqueID,
		VaaID:            v.MessageID(),
		Timestamp:        &v.Timestamp,
		Version:          v.Version,
		EmitterChain:     v.EmitterChain,
		EmitterAddr:      v.EmitterAddress.String(),
		Sequence:         strconv.FormatUint(v.Sequence, 10),
		GuardianSetIndex: v.GuardianSetIndex,
		Vaa:              serializedVaa,
		UpdatedAt:        &t,
		ConsistencyLevel: v.ConsistencyLevel,
		Digest:           utils.NormalizeHex(v.HexDigest()),
	}
}
