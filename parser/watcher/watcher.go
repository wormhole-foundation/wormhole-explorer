package watcher

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// Watcher represents a listener of database changes.
type Watcher struct {
	db      *mongo.Database
	dbName  string
	handler WatcherFunc
	logger  *zap.Logger
}

// WatcherFunc is a function to send database changes.
type WatcherFunc func(*Event)

type watchEvent struct {
	DocumentKey    documentKey `bson:"documentKey"`
	OperationType  string      `bson:"operationType"`
	DbFullDocument Event       `bson:"fullDocument"`
}
type documentKey struct {
	ID string `bson:"_id"`
}

// Event represents a database change.
type Event struct {
	ID   string `bson:"_id"`
	Vaas []byte
}

const queryTemplate = `
	[
		{ 
			"$match" : {
				"operationType" : "insert",
				"ns": { "$in": [{"db": "%s", "coll": "vaas"}] } 
			}
		}
   	]
`

// NewWatcher creates a new database event watcher.
func NewWatcher(db *mongo.Database, dbName string, handler WatcherFunc, logger *zap.Logger) *Watcher {
	return &Watcher{
		db:      db,
		dbName:  dbName,
		handler: handler,
		logger:  logger,
	}
}

// Start executes database event consumption.
func (w *Watcher) Start(ctx context.Context) error {
	query := fmt.Sprintf(queryTemplate, w.dbName, w.dbName)
	var steps []bson.D
	err := bson.UnmarshalExtJSON([]byte(query), true, &steps)
	if err != nil {
		return err
	}

	stream, err := w.db.Watch(ctx, steps)
	if err != nil {
		return err
	}
	go func() {
		for stream.Next(ctx) {
			var e watchEvent
			if err := stream.Decode(&e); err != nil {
				w.logger.Error("Error unmarshalling event", zap.Error(err))
				continue
			}
			w.handler(&Event{
				ID:   e.DbFullDocument.ID,
				Vaas: e.DbFullDocument.Vaas,
			})
		}
	}()
	return nil
}
