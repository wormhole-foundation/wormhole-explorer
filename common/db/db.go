package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	connectTimeout = 10 * time.Second
)

// Session is a plain-old-data struct that represents a handle to a MongoDB database.
type Session struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// Connect to a MongoDB database.
//
// Returns a struct that represents a handle to the database.
//
// Most of the time, you probably want to defer a call to `DB.Disconnect()`
// after calling this function.
func Connect(ctx context.Context, uri, databaseName string) (*Session, error) {

	// Create a timed sub-context for the connection attempt
	subContext, cancelFunc := context.WithTimeout(ctx, connectTimeout)
	defer cancelFunc()

	// Connect to MongoDB
	client, err := mongo.Connect(subContext, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to make sure we're actually connected
	//
	// This can detect a misconfuiguration error when a service is being initialized,
	// rather than waiting for the first query to fail in the service's processing loop.
	err = client.Ping(subContext, readpref.Primary())
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB database: %w", err)
	}

	// Populate the result struct and return
	db := &Session{
		Client:   client,
		Database: client.Database(databaseName),
	}
	return db, nil
}

// Disconnect from a MongoDB database.
func (db *Session) DisconnectWithTimeout(timeout time.Duration) error {

	// Create a timed sub-context for the disconnection attempt
	subContext, cancelFunc := context.WithTimeout(context.Background(), timeout)
	defer cancelFunc()

	// Attempt to disconnect
	err := db.Client.Disconnect(subContext)
	if err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}

	return nil
}
