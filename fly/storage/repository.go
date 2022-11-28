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
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// TODO separate and maybe share between fly and web
type Repository struct {
	db          *mongo.Database
	log         *zap.Logger
	collections struct {
		vaas           *mongo.Collection
		heartbeats     *mongo.Collection
		observations   *mongo.Collection
		governorConfig *mongo.Collection
		governorStatus *mongo.Collection
		vaasPythnet    *mongo.Collection
		vaaCounts      *mongo.Collection
	}
}

// TODO wrap repository with a service that filters using redis
func NewRepository(db *mongo.Database, log *zap.Logger) *Repository {
	return &Repository{db, log, struct {
		vaas           *mongo.Collection
		heartbeats     *mongo.Collection
		observations   *mongo.Collection
		governorConfig *mongo.Collection
		governorStatus *mongo.Collection
		vaasPythnet    *mongo.Collection
		vaaCounts      *mongo.Collection
	}{
		vaas:           db.Collection("vaas"),
		heartbeats:     db.Collection("heartbeats"),
		observations:   db.Collection("observations"),
		governorConfig: db.Collection("governorConfig"),
		governorStatus: db.Collection("governorStatus"),
		vaasPythnet:    db.Collection("vaasPythnet"),
		vaaCounts:      db.Collection("vaaCounts")}}
}

func (s *Repository) UpsertVaa(ctx context.Context, v *vaa.VAA, serializedVaa []byte) error {
	id := v.MessageID()
	now := time.Now()
	vaaDoc := VaaUpdate{
		ID:               v.MessageID(),
		Timestamp:        &v.Timestamp,
		Version:          v.Version,
		EmitterChain:     v.EmitterChain,
		EmitterAddr:      v.EmitterAddress.String(),
		Sequence:         v.Sequence,
		GuardianSetIndex: v.GuardianSetIndex,
		Vaa:              serializedVaa,
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
	if vaa.ChainIDPythNet == v.EmitterChain {
		result, err = s.collections.vaasPythnet.UpdateByID(ctx, id, update, opts)
	} else {
		result, err = s.collections.vaas.UpdateByID(ctx, id, update, opts)
	}
	if err == nil && s.isNewRecord(result) {
		s.updateVAACount(v.EmitterChain)
	}
	return err
}

func (s *Repository) UpsertObservation(o *gossipv1.SignedObservation) error {
	vaaID := strings.Split(o.MessageId, "/")
	chainIdStr, emitter, sequenceStr := vaaID[0], vaaID[1], vaaID[2]
	id := fmt.Sprintf("%s/%s/%s", o.MessageId, hex.EncodeToString(o.Addr), hex.EncodeToString(o.Hash))
	now := time.Now()
	//TODO error handling
	chainId, err := strconv.ParseUint(chainIdStr, 10, 16)
	sequence, err := strconv.ParseUint(sequenceStr, 10, 64)
	addr := eth_common.BytesToAddress(o.GetAddr())
	obs := ObservationUpdate{
		ChainID:      vaa.ChainID(chainId),
		Emitter:      emitter,
		Sequence:     sequence,
		MessageID:    o.GetMessageId(),
		Hash:         o.GetHash(),
		TxHash:       o.GetTxHash(),
		GuardianAddr: addr.String(),
		Signature:    o.GetSignature(),
		UpdatedAt:    &now,
	}

	update := bson.M{
		"$set":         obs,
		"$setOnInsert": indexedAt(now),
	}
	opts := options.Update().SetUpsert(true)
	_, err = s.collections.observations.UpdateByID(context.TODO(), id, update, opts)
	if err != nil {
		s.log.Error("Error inserting observation", zap.Error(err))
	}
	return err
}

func (s *Repository) UpsertHeartbeat(hb *gossipv1.Heartbeat) error {
	id := hb.GuardianAddr
	now := time.Now()
	update := bson.D{{Key: "$set", Value: hb}, {Key: "$set", Value: bson.D{{Key: "updatedAt", Value: now}}}, {Key: "$setOnInsert", Value: bson.D{{Key: "indexedAt", Value: now}}}}
	opts := options.Update().SetUpsert(true)
	_, err := s.collections.heartbeats.UpdateByID(context.TODO(), id, update, opts)
	return err
}

func (s *Repository) UpsertGovernorConfig(govC *gossipv1.SignedChainGovernorConfig) error {
	id := hex.EncodeToString(govC.GuardianAddr)
	now := time.Now()
	var cfg gossipv1.ChainGovernorConfig
	err := proto.Unmarshal(govC.Config, &cfg)
	if err != nil {
		s.log.Error("Error unmarshalling govr config", zap.Error(err))
		return err
	}
	update := bson.D{{Key: "$set", Value: govC}, {Key: "$set", Value: bson.D{{Key: "parsedConfig", Value: cfg}}}, {Key: "$set", Value: bson.D{{Key: "updatedAt", Value: now}}}, {Key: "$setOnInsert", Value: bson.D{{Key: "createdAt", Value: now}}}}
	opts := options.Update().SetUpsert(true)
	_, err2 := s.collections.governorConfig.UpdateByID(context.TODO(), id, update, opts)

	if err2 != nil {
		s.log.Error("Error inserting govr cfg", zap.Error(err2))
	}
	return err2
}

func (s *Repository) UpsertGovernorStatus(govS *gossipv1.SignedChainGovernorStatus) error {
	id := hex.EncodeToString(govS.GuardianAddr)
	now := time.Now()
	var status gossipv1.ChainGovernorStatus
	err := proto.Unmarshal(govS.Status, &status)
	if err != nil {
		s.log.Error("Error unmarshalling govr status", zap.Error(err))
		return err
	}
	update := bson.D{{Key: "$set", Value: govS}, {Key: "$set", Value: bson.D{{Key: "parsedStatus", Value: status}}}, {Key: "$set", Value: bson.D{{Key: "updatedAt", Value: now}}}, {Key: "$setOnInsert", Value: bson.D{{Key: "createdAt", Value: now}}}}

	opts := options.Update().SetUpsert(true)
	_, err2 := s.collections.governorStatus.UpdateByID(context.TODO(), id, update, opts)

	if err2 != nil {
		s.log.Error("Error inserting govr status", zap.Error(err2))
	}
	return err2
}

func (s *Repository) updateVAACount(chainID vaa.ChainID) {
	update := bson.D{{Key: "$inc", Value: bson.D{{Key: "count", Value: 1}}}}
	opts := options.Update().SetUpsert(true)
	_, _ = s.collections.vaaCounts.UpdateByID(context.TODO(), chainID, update, opts)
}

func (s *Repository) isNewRecord(result *mongo.UpdateResult) bool {
	return result.MatchedCount == 0 && result.ModifiedCount == 0 && result.UpsertedCount == 1
}

// GetMongoStatus get mongo server status
func (r *Repository) GetMongoStatus(ctx context.Context) (*MongoStatus, error) {
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
