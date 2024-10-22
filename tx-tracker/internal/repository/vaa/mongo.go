package vaa

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type RepositoryMongoDB struct {
	db                 *mongo.Database
	logger             *zap.Logger
	vaas               *mongo.Collection
	globalTransactions *mongo.Collection
}

// NewMongoVaaRepository create a new VaaRepositoryMongoDB.
func NewMongoVaaRepository(db *mongo.Database, logger *zap.Logger) *RepositoryMongoDB {
	return &RepositoryMongoDB{db: db,
		logger:             logger.With(zap.String("module", "VaaRepository")),
		vaas:               db.Collection("vaas"),
		globalTransactions: db.Collection("globalTransactions"),
	}
}

func (r *RepositoryMongoDB) FindById(ctx context.Context, id string) (*VaaDoc, error) {
	var vaaDoc VaaDoc
	err := r.vaas.FindOne(ctx, bson.M{"_id": id}).Decode(&vaaDoc)
	return &vaaDoc, err
}

func (r *RepositoryMongoDB) GetVaa(ctx context.Context, id string) (*VaaDoc, error) {
	return r.FindById(ctx, id)
}

// GetTxHash method to support
func (r *RepositoryMongoDB) GetTxHash(ctx context.Context, vaaDigest string) (string, error) {
	vaaDoc, err := r.FindById(ctx, vaaDigest)
	if err != nil {
		return "", err
	}
	return vaaDoc.TxHash, nil
}
