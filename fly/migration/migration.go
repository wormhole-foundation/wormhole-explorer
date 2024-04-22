package migration

import (
	"context"
	"errors"

	"github.com/wormhole-foundation/wormhole-explorer/common/repository"
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
	err = db.CreateCollection(context.TODO(), repository.VaaIdTxHash)
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// Create duplicateVaas collection.
	err = db.CreateCollection(context.TODO(), repository.DuplicateVaas)
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

	indexVaaByEmitteChainEmitterAddrSequence := mongo.IndexModel{
		Keys: bson.D{
			{Key: "emitterChain", Value: 1},
			{Key: "emitterAddr", Value: 1},
			{Key: "sequence", Value: 1},
		}}
	_, err = db.Collection("vaas").Indexes().CreateOne(context.TODO(), indexVaaByEmitteChainEmitterAddrSequence)
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	indexVaaByTimestampId := mongo.IndexModel{
		Keys: bson.D{
			{Key: "timestamp", Value: -1},
			{Key: "_id", Value: -1},
		}}
	_, err = db.Collection("vaas").Indexes().CreateOne(context.TODO(), indexVaaByTimestampId)
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	indexVaaByTxHash := mongo.IndexModel{
		Keys: bson.D{
			{Key: "txHash", Value: 1},
		}}
	_, err = db.Collection("vaas").Indexes().CreateOne(context.TODO(), indexVaaByTxHash)
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
	_, err = db.Collection(repository.VaaIdTxHash).Indexes().CreateOne(context.TODO(), indexVaaIdTxHashByTxHash)
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// create index in globalTransactions collect.
	indexGlobalTransactionsByOriginTx := mongo.IndexModel{
		Keys: bson.D{{Key: "originTx.from", Value: 1}}}
	_, err = db.Collection("globalTransactions").Indexes().CreateOne(context.TODO(), indexGlobalTransactionsByOriginTx)
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// create index in globalTransactions collection for wormchain nested txHash.
	indexGlobalTransactionsByOriginTxInAttribute := mongo.IndexModel{
		Keys: bson.D{{Key: "originTx.attribute.value.originTxHash", Value: 1}}}
	_, err = db.Collection("globalTransactions").Indexes().CreateOne(context.TODO(), indexGlobalTransactionsByOriginTxInAttribute)
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// create index in globalTransactions collection by originTx.attribute.value.originAddress.
	indexGlobalTransactionsByOriginTxInAttributeOriginAddress := mongo.IndexModel{
		Keys: bson.D{{Key: "originTx.attribute.value.originAddress", Value: 1}}}
	_, err = db.Collection("globalTransactions").Indexes().CreateOne(context.TODO(), indexGlobalTransactionsByOriginTxInAttributeOriginAddress)
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// create index in globalTransactions collection by originTx nativeTxHash.
	indexGlobalTransactionsByOriginNativeTxHash := mongo.IndexModel{
		Keys: bson.D{{Key: "originTx.nativeTxHash", Value: 1}}}
	_, err = db.Collection("globalTransactions").Indexes().CreateOne(context.TODO(), indexGlobalTransactionsByOriginNativeTxHash)
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// create index in globalTransactions collection by destination txHash.
	indexGlobalTransactionsByDestinationTxHash := mongo.IndexModel{
		Keys: bson.D{{Key: "destinationTx.txHash", Value: 1}}}
	_, err = db.Collection("globalTransactions").Indexes().CreateOne(context.TODO(), indexGlobalTransactionsByDestinationTxHash)
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// create index in globalTransactions collection by timestamp/_id sort.
	indexGlobalTransactionsByTimestampAndId := mongo.IndexModel{
		Keys: bson.D{{Key: "originTx.timestamp", Value: -1}, {Key: "_id", Value: -1}}}
	_, err = db.Collection("globalTransactions").Indexes().CreateOne(context.TODO(), indexGlobalTransactionsByTimestampAndId)
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// create index in parsedVaa collection by standardizedProperties toAddress.
	indexParsedVaaByStandardizedPropertiesToAddress := mongo.IndexModel{
		Keys: bson.D{{Key: "standardizedProperties.toAddress", Value: 1}}}
	_, err = db.Collection("parsedVaa").Indexes().CreateOne(context.TODO(), indexParsedVaaByStandardizedPropertiesToAddress)
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// create index in parsedVaa collection by parsedPayload tokenAddress.
	indexParsedVaaByParsedPayloadTokenAddress := mongo.IndexModel{
		Keys: bson.D{{Key: "parsedPayload.tokenAddress", Value: 1}}}
	_, err = db.Collection("parsedVaa").Indexes().CreateOne(context.TODO(), indexParsedVaaByParsedPayloadTokenAddress)
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// create index in parsedVaa collection by indexedAt.
	indexParsedVaaByIndexedAt := mongo.IndexModel{
		Keys: bson.D{{Key: "indexedAt", Value: 1}}}
	_, err = db.Collection("parsedVaa").Indexes().CreateOne(context.TODO(), indexParsedVaaByIndexedAt)
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// create index in parsedVaa collection by rawStandardizedProperties.appIds.
	indexParsedVaaRawByAppIds := mongo.IndexModel{
		Keys: bson.D{{Key: "rawStandardizedProperties.appIds", Value: 1},
			{Key: "timestamp", Value: -1},
			{Key: "_id", Value: -1},
		}}

	_, err = db.Collection("parsedVaa").Indexes().CreateOne(context.TODO(), indexParsedVaaRawByAppIds)
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// create index for querying by fromChain
	compoundIndexParsedVaaByFromToChain := mongo.IndexModel{
		Keys: bson.D{
			{Key: "rawStandardizedProperties.fromChain", Value: -1},
			{Key: "timestamp", Value: -1},
			{Key: "_id", Value: -1},
		}}
	_, err = db.Collection("parsedVaa").Indexes().CreateOne(context.TODO(), compoundIndexParsedVaaByFromToChain)
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// create index for querying by toChain
	indexParsedVaaByToChain := mongo.IndexModel{
		Keys: bson.D{
			{Key: "rawStandardizedProperties.toChain", Value: 1},
			{Key: "timestamp", Value: -1},
			{Key: "_id", Value: -1},
		},
	}
	_, err = db.Collection("parsedVaa").Indexes().CreateOne(context.TODO(), indexParsedVaaByToChain)
	if err != nil && isNotAlreadyExistsError(err) {
		return err
	}

	// create index for duplicateVaas by vaaId
	indexDuplicateVaasByVaadID := mongo.IndexModel{
		Keys: bson.D{
			{Key: "vaaId", Value: -1},
		},
	}
	_, err = db.Collection(repository.DuplicateVaas).Indexes().CreateOne(context.TODO(), indexDuplicateVaasByVaadID)
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
