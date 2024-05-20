package vaa

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type Repository struct {
	db                 *mongo.Database
	logger             *zap.Logger
	vaas               *mongo.Collection
	globalTransactions *mongo.Collection
}

type VaaDoc struct {
	ID     string `bson:"_id" json:"id"`
	Vaa    []byte `bson:"vaas" json:"vaa"`
	TxHash string `bson:"txHash" json:"txHash"`
}

// NewRepository create a new Repository.
func NewRepository(db *mongo.Database, logger *zap.Logger) *Repository {
	return &Repository{db: db,
		logger:             logger.With(zap.String("module", "VaaRepository")),
		vaas:               db.Collection("vaas"),
		globalTransactions: db.Collection("globalTransactions"),
	}
}

func (r *Repository) FindById(ctx context.Context, id string) (*VaaDoc, error) {
	var vaaDoc VaaDoc
	err := r.vaas.FindOne(ctx, bson.M{"_id": id}).Decode(&vaaDoc)
	return &vaaDoc, err
}
