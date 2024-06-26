package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type GuardianSetRepository struct {
	db          *mongo.Database
	logger      *zap.Logger
	guardianSet *mongo.Collection
}

// GuardianSetKeyDoc is a key document for GuardianSet.
type GuardianSetKeyDoc struct {
	Index   uint32 `bson:"index" json:"index"`
	Address []byte `bson:"address" json:"address"`
}

// GuardianSetDoc is a document for GuardianSet.
type GuardianSetDoc struct {
	GuardianSetIndex uint32              `bson:"_id" json:"guardianSetIndex"`
	Keys             []GuardianSetKeyDoc `bson:"keys" json:"keys"`
	ExpirationTime   *time.Time          `bson:"expirationTime" json:"expirationTime"`
	UpdatedAt        time.Time           `bson:"updatedAt"`
}

// NewGuardianSetRepository create a new guardian set repository.
func NewGuardianSetRepository(db *mongo.Database, logger *zap.Logger) *GuardianSetRepository {
	return &GuardianSetRepository{db: db,
		logger:      logger.With(zap.String("module", "GuardianSetRepository")),
		guardianSet: db.Collection(GuardianSets),
	}
}

// Upsert upserts a guardian set document.
func (r *GuardianSetRepository) Upsert(ctx context.Context, doc *GuardianSetDoc) error {
	now := time.Now()
	update := bson.M{
		"$set":         doc,
		"$setOnInsert": IndexedAt(now),
		"$inc":         bson.D{{Key: "revision", Value: 1}},
	}
	opts := options.Update().SetUpsert(true)
	_, err := r.guardianSet.UpdateByID(ctx, doc.GuardianSetIndex, update, opts)
	return err
}

// FindByIndex finds guardian set by index.
func (r *GuardianSetRepository) FindByIndex(ctx context.Context, index uint32) (*GuardianSetDoc, error) {
	var guardianSetDoc GuardianSetDoc
	err := r.guardianSet.FindOne(ctx, bson.M{"_id": index}).Decode(&guardianSetDoc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &guardianSetDoc, err
}

// FindAll finds all guardian sets.
func (r *GuardianSetRepository) FindAll(ctx context.Context) ([]*GuardianSetDoc, error) {
	cursor, err := r.guardianSet.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var guardianSetDocs []*GuardianSetDoc
	for cursor.Next(ctx) {
		var guardianSetDoc GuardianSetDoc
		if err := cursor.Decode(&guardianSetDoc); err != nil {
			return nil, err
		}
		guardianSetDocs = append(guardianSetDocs, &guardianSetDoc)
	}
	return guardianSetDocs, nil
}
