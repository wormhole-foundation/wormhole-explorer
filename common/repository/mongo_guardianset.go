package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// MongoGuardianSetRepository repository.
type MongoGuardianSetRepository struct {
	db          *mongo.Database
	logger      *zap.Logger
	guardianSet *mongo.Collection
}

// GuardianSetKey is a key document for GuardianSet.
type GuardianSetKey struct {
	Index   uint32 `bson:"index" json:"index" db:"index"`
	Address []byte `bson:"address" json:"address" db:"address"`
}

// GuardianSet is a document for GuardianSet.
type GuardianSet struct {
	GuardianSetIndex uint32           `bson:"_id" json:"guardianSetIndex" db:"guardian_set_id"`
	Keys             []GuardianSetKey `bson:"keys" json:"keys" db:"keys"`
	ExpirationTime   *time.Time       `bson:"expirationTime" json:"expirationTime" db:"expiration_time"`
	CreatedAt        *time.Time       `bson:"-" json:"createdAt" db:"created_at"`
	UpdatedAt        time.Time        `bson:"updatedAt" json:"updatedAt" db:"updated_at"`
}

// NewMongoGuardianSetRepository create a new guardian set repository.
func NewMongoGuardianSetRepository(db *mongo.Database, logger *zap.Logger) *MongoGuardianSetRepository {
	return &MongoGuardianSetRepository{db: db,
		logger:      logger.With(zap.String("module", "GuardianSetRepository")),
		guardianSet: db.Collection(GuardianSets),
	}
}

// Upsert upserts a guardian set document.
func (r *MongoGuardianSetRepository) Upsert(ctx context.Context, doc *GuardianSet) error {
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
func (r *MongoGuardianSetRepository) FindByIndex(ctx context.Context, index uint32) (*GuardianSet, error) {
	var guardianSetDoc GuardianSet
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
func (r *MongoGuardianSetRepository) FindAll(ctx context.Context) ([]*GuardianSet, error) {
	cursor, err := r.guardianSet.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var guardianSetDocs []*GuardianSet
	for cursor.Next(ctx) {
		var guardianSetDoc GuardianSet
		if err := cursor.Decode(&guardianSetDoc); err != nil {
			return nil, err
		}
		guardianSetDocs = append(guardianSetDocs, &guardianSetDoc)
	}
	return guardianSetDocs, nil
}
