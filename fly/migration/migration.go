package migration

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TODO: move this to migration tool that support mongodb.
func Run(db *mongo.Database) error {
	// Created governorConfig collection.
	err := db.CreateCollection(context.TODO(), "governorConfig")
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// Created governorStatus collection.
	err = db.CreateCollection(context.TODO(), "governorStatus")
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// Created heartbeats collection.
	err = db.CreateCollection(context.TODO(), "heartbeats")
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// Created observations collection.
	err = db.CreateCollection(context.TODO(), "observations")
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// Created vaaCounts collection.
	err = db.CreateCollection(context.TODO(), "vaaCounts")
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// Create vaas collection.
	err = db.CreateCollection(context.TODO(), "vaas")
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// Create vassPythnet capped collection.
	isCapped := true
	var sizeCollection, maxDocuments int64 = 50 * 1024 * 1024, 10000
	collectionOptions := options.CreateCollectionOptions{
		Capped:       &isCapped,
		SizeInBytes:  &sizeCollection,
		MaxDocuments: &maxDocuments}
	err = db.CreateCollection(context.TODO(), "vaasPythnet", &collectionOptions)
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// Create vaaIdTxHash collection.
	err = db.CreateCollection(context.TODO(), "vaaIdTxHash")
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// create index in vaas collection by vaa key (emitterchain, emitterAddr, sequence)
	indexVaaByKey := mongo.IndexModel{
		Keys: bson.D{
			{Key: "timestamp", Value: -1},
			{Key: "emitterAddr", Value: 1},
			{Key: "emitterChain", Value: 1},
		}}
	_, err = db.Collection("vaas").Indexes().CreateOne(context.TODO(), indexVaaByKey)
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	indexVaaByTimestamp := mongo.IndexModel{
		Keys: bson.D{
			{Key: "emitterChain", Value: 1},
			{Key: "emitterAddr", Value: 1},
			{Key: "sequence", Value: 1},
		}}
	_, err = db.Collection("vaas").Indexes().CreateOne(context.TODO(), indexVaaByTimestamp)
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// create index in observations collection by indexedAt.
	indexObservationsByIndexedAt := mongo.IndexModel{Keys: bson.D{{Key: "indexedAt", Value: 1}}}
	_, err = db.Collection("observations").Indexes().CreateOne(context.TODO(), indexObservationsByIndexedAt)
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// create index in observations collect.
	indexObservationsByEmitterChainAndAddressAndSequence := mongo.IndexModel{
		Keys: bson.D{
			{Key: "emitterChain", Value: 1},
			{Key: "emitterAddr", Value: 1},
			{Key: "sequence", Value: 1}}}
	_, err = db.Collection("observations").Indexes().CreateOne(context.TODO(), indexObservationsByEmitterChainAndAddressAndSequence)
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// create index in vaaIdTxHash collect.
	indexVaaIdTxHashByTxHash := mongo.IndexModel{
		Keys: bson.D{{Key: "txHash", Value: 1}}}
	_, err = db.Collection("vaaIdTxHash").Indexes().CreateOne(context.TODO(), indexVaaIdTxHashByTxHash)
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// create index in globaltransactions collect.
	indexGlobalTransactionsByOriginTx := mongo.IndexModel{
		Keys: bson.D{{Key: "originTx.from", Value: 1}}}
	_, err = db.Collection("globaltransactions").Indexes().CreateOne(context.TODO(), indexGlobalTransactionsByOriginTx)
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
