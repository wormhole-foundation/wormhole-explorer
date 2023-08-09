package migration

import (
	"context"
	"errors"

	"github.com/wormhole-foundation/wormhole-explorer/parser/parser"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TODO: move this to migration tool that support mongodb.
func Run(db *mongo.Database) error {
	// Created parsedVaa collection.
	err := db.CreateCollection(context.TODO(), parser.ParsedVAACollection)
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// create index in observations collection by indexedAt.
	indexToAddress := mongo.IndexModel{Keys: bson.D{{Key: "standardizedProperties.toAddress", Value: 1}}}
	_, err = db.Collection(parser.ParsedVAACollection).Indexes().CreateOne(context.TODO(), indexToAddress)
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	return nil
}

func isNotAlreadyExistsError(err error) bool {
	target := &mongo.CommandError{}
	isCommandError := errors.As(err, target)
	if !isCommandError || err.(mongo.CommandError).Code != 48 {
		return true
	}
	return false
}
