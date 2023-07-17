package vaa

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2/bson"
)

// Repository repository struct definition.
type Repository struct {
	db     *mongo.Database
	logger *zap.Logger
	vaas   *mongo.Collection
}

// VaaDoc vaa document struct definition.
type VaaDoc struct {
	ID  string `bson:"_id" json:"id"`
	Vaa []byte `bson:"vaas" json:"vaa"`
}

// NewRepository create a new Repository.
func NewRepository(db *mongo.Database, logger *zap.Logger) *Repository {
	return &Repository{db: db,
		logger: logger.With(zap.String("module", "VaaRepository")),
		vaas:   db.Collection("vaas"),
	}
}

// FindById find a vaa by id.
func (r *Repository) FindById(ctx context.Context, id string) (*VaaDoc, error) {
	var vaaDoc VaaDoc
	err := r.vaas.FindOne(ctx, bson.M{"_id": id}).Decode(&vaaDoc)
	return &vaaDoc, err
}
