package storage

import (
	"context"
	"encoding/hex"
	"fmt"
	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
	"github.com/certusone/wormhole/node/pkg/vaa"
	eth_common "github.com/ethereum/go-ethereum/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

// TODO separate and maybe share between fly and web
type Repository struct {
	db          *mongo.Database
	log         *zap.Logger
	collections struct {
		vaas         *mongo.Collection
		invalidVaas  *mongo.Collection
		heartbeats   *mongo.Collection
		observations *mongo.Collection
	}
}

// TODO wrap repository with a service that filters using redis
func NewRepository(db *mongo.Database, log *zap.Logger) *Repository {
	return &Repository{db, log, struct {
		vaas         *mongo.Collection
		invalidVaas  *mongo.Collection
		heartbeats   *mongo.Collection
		observations *mongo.Collection
	}{vaas: db.Collection("vaas"), invalidVaas: db.Collection("invalid_vaas"), heartbeats: db.Collection("heartbeats"), observations: db.Collection("observations")}}
}

func (s *Repository) UpsertVaa(v *vaa.VAA, serializedVaa []byte) error {
	return s.upsertVaa(v, serializedVaa, true)
}

func (s *Repository) UpsertInvalidVaa(v *vaa.VAA, serializedVaa []byte) error {
	return s.upsertVaa(v, serializedVaa, false)
}

func (s *Repository) upsertVaa(v *vaa.VAA, serializedVaa []byte, isValid bool) error {
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
	}

	opts := options.Update().SetUpsert(true)
	var err error
	if isValid {
		_, err = s.collections.vaas.UpdateByID(context.TODO(), id, update, opts)
	} else {
		_, err = s.collections.invalidVaas.UpdateByID(context.TODO(), id, update, opts)
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
	return nil
}

func (s *Repository) UpsertHeartbeat(hb *gossipv1.Heartbeat) error {
	id := hb.GuardianAddr
	now := time.Now()
	update := bson.D{{Key: "$set", Value: hb}, {Key: "$set", Value: bson.D{{Key: "updatedAt", Value: now}}}, {Key: "$setOnInsert", Value: bson.D{{Key: "indexedAt", Value: now}}}}
	opts := options.Update().SetUpsert(true)
	_, err := s.collections.heartbeats.UpdateByID(context.TODO(), id, update, opts)
	return err
}
