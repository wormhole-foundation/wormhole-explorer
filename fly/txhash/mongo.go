package txhash

import (
	"context"
	"fmt"
	"time"

	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
	"github.com/wormhole-foundation/wormhole-explorer/common/repository"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type mongoTxHash struct {
	vaaIdTxHashCollection *mongo.Collection
	logger                *zap.Logger
}

type vaaIdTxHashUpdate struct {
	ChainID   sdk.ChainID `bson:"emitterChain"`
	Emitter   string      `bson:"emitterAddr"`
	Sequence  string      `bson:"sequence"`
	TxHash    string      `bson:"txHash"`
	UpdatedAt *time.Time  `bson:"updatedAt"`
}

func NewMongoTxHash(database *mongo.Database, logger *zap.Logger) *mongoTxHash {
	return &mongoTxHash{
		vaaIdTxHashCollection: database.Collection(repository.VaaIdTxHash),
		logger:                logger,
	}
}

func (m *mongoTxHash) Set(ctx context.Context, vaaID string, txHash TxHash) error {
	id := fmt.Sprintf("%d/%s/%s", txHash.ChainID, txHash.Emitter, txHash.Sequence)
	now := time.Now()
	udpate := vaaIdTxHashUpdate{
		ChainID:   txHash.ChainID,
		Emitter:   txHash.Emitter,
		Sequence:  txHash.Sequence,
		TxHash:    txHash.TxHash,
		UpdatedAt: &now,
	}

	updateVaaTxHash := bson.M{
		"$set":         udpate,
		"$setOnInsert": repository.IndexedAt(now),
		"$inc":         bson.D{{Key: "revision", Value: 1}},
	}
	_, err := m.vaaIdTxHashCollection.UpdateByID(ctx, id, updateVaaTxHash, options.Update().SetUpsert(true))
	if err != nil {
		m.logger.Error("Error inserting vaaIdTxHash in mongodb", zap.String("id", id), zap.Error(err))
		return err
	}
	return nil
}

func (r *mongoTxHash) SetObservation(ctx context.Context, o *gossipv1.SignedObservation) error {
	txHash, err := CreateTxHash(r.logger, o)
	if err != nil {
		r.logger.Error("Error creating txHash", zap.Error(err))
		return err
	}
	return r.Set(ctx, o.MessageId, *txHash)
}

func (m *mongoTxHash) Get(ctx context.Context, vaaID string) (*TxHash, error) {
	var mongoTxHash TxHash
	if err := m.vaaIdTxHashCollection.FindOne(ctx, bson.M{"_id": vaaID}).Decode(&mongoTxHash); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrTxHashNotFound
		}
		m.logger.Error("Finding vaaIdTxHash", zap.String("id", vaaID), zap.Error(err))
		return nil, err
	}
	return &mongoTxHash, nil
}

func (r *mongoTxHash) GetName() string {
	return "mongo"
}
