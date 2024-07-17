//go:generate mockgen -source=repository.go -destination=mocks/repository.go -package=mocks
package pipeline

import (
	"context"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/repository"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// Interface
type IRepository interface {
	GetVaaIdTxHash(ctx context.Context, id string) (*VaaIdTxHash, error)
	UpdateVaaDocTxHash(ctx context.Context, id string, txhash string) error
}

// Repository is the repository data access layer.
type Repository struct {
	db          *mongo.Database
	log         *zap.Logger
	collections struct {
		vaas        *mongo.Collection
		vaaIdTxHash *mongo.Collection
	}
}

// NewRepository creates a new repository.
func NewRepository(db *mongo.Database, log *zap.Logger) *Repository {
	return &Repository{db, log, struct {
		vaas        *mongo.Collection
		vaaIdTxHash *mongo.Collection
	}{
		vaas:        db.Collection(repository.Vaas),
		vaaIdTxHash: db.Collection(repository.VaaIdTxHash),
	}}
}

// VaaIdTxHash represents a vaaIdTxHash document.
type VaaIdTxHash struct {
	VaaID     string      `bson:"vaaId"`
	ChainID   vaa.ChainID `bson:"emitterChain"`
	Emitter   string      `bson:"emitterAddr"`
	Sequence  string      `bson:"sequence"`
	TxHash    string      `bson:"txHash"`
	UpdatedAt *time.Time  `bson:"updatedAt"`
}

// GetVaaIdTxHash returns a vaaIdTxHash document.
func (r *Repository) GetVaaIdTxHash(ctx context.Context, id string) (*VaaIdTxHash, error) {
	var v VaaIdTxHash
	err := r.collections.vaaIdTxHash.FindOne(ctx, bson.M{"_id": id}).Decode(&v)
	return &v, err
}

// VaaUpdate represents a vaa document.
type VaaUpdate struct {
	TxHash    string     `bson:"txHash,omitempty"`
	UpdatedAt *time.Time `bson:"updatedAt"`
}

// UpdateVaaTxHash update a txhash in a vaa document.
func (r *Repository) UpdateVaaDocTxHash(ctx context.Context, id string, txhash string) error {
	vaaDoc := &VaaUpdate{
		TxHash:    txhash,
		UpdatedAt: &time.Time{},
	}

	update := bson.M{
		"$set": vaaDoc,
		"$inc": bson.D{{Key: "revision", Value: 1}},
	}

	_, err := r.collections.vaas.UpdateByID(ctx, id, update, nil)
	return err
}
