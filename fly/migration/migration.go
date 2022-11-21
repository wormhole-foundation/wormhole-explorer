package migration

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Run(db *mongo.Database) error {
	// TODO: change this to use a migration tool.
	isCapped := true
	var sizeCollection, maxDocuments int64 = 50 * 1024 * 1024, 10000
	collectionOptions := options.CreateCollectionOptions{
		Capped:       &isCapped,
		SizeInBytes:  &sizeCollection,
		MaxDocuments: &maxDocuments}
	err := db.CreateCollection(context.TODO(), "vaasPythnet", &collectionOptions)
	if err != nil {
		target := &mongo.CommandError{}
		isCommandError := errors.As(err, target)
		if !isCommandError || err.(mongo.CommandError).Code != 48 {
			return err
		}
	}
	return nil
}
