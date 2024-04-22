package watcher

import (
	"context"
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	pipelineAlert "github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/alert"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/metrics"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// Watcher represents a listener of database changes.
type Watcher struct {
	db          *mongo.Database
	dbName      string
	handler     WatcherFunc
	alertClient alert.AlertClient
	metrics     metrics.Metrics
	logger      *zap.Logger
}

// WatcherFunc is a function to send database changes.
type WatcherFunc func(context.Context, *Event)

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
	ID               string     `bson:"_id"`
	ChainID          uint16     `bson:"emitterChain"`
	EmitterAddress   string     `bson:"emitterAddr"`
	Sequence         string     `bson:"sequence"`
	GuardianSetIndex uint32     `bson:"guardianSetIndex"`
	Vaa              []byte     `bson:"vaas"`
	IndexedAt        time.Time  `bson:"indexedAt"`
	Timestamp        *time.Time `bson:"timestamp"`
	UpdatedAt        *time.Time `bson:"updatedAt"`
	TxHash           string     `bson:"txHash"`
	Version          uint16     `bson:"version"`
	Revision         uint16     `bson:"revision"`
	Hash             []byte     `bson:"hash"`
	IsDuplicated     bool       `bson:"isDuplicated"`
}

const queryTemplate = `
	[
		{ 
			"$match" : {
				"operationType" : "insert",
				"ns": { "$in": [{"db": "%s", "coll": "vaasPythnet"}, {"db": "%s", "coll": "vaas"}] } 
			}
		}
   	]
`

// NewWatcher creates a new database event watcher.
func NewWatcher(ctx context.Context, db *mongo.Database, dbName string, handler WatcherFunc, alertClient alert.AlertClient, metrics metrics.Metrics, logger *zap.Logger) *Watcher {
	return &Watcher{
		db:          db,
		dbName:      dbName,
		handler:     handler,
		metrics:     metrics,
		alertClient: alertClient,
		logger:      logger,
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
				alertContext := alert.AlertContext{
					Details: e.toMapAlertDetail(),
					Error:   err,
				}
				w.alertClient.CreateAndSend(ctx, pipelineAlert.ErrorDecodeWatcherEvent, alertContext)
				continue
			}
			w.metrics.IncVaaFromMongoStream(e.DbFullDocument.ChainID)
			w.handler(ctx, &e.DbFullDocument)
		}
	}()
	return nil
}

// toAlertDetail returns from the watch event an map with the alert details.
func (e *watchEvent) toMapAlertDetail() map[string]string {
	detail := make(map[string]string)
	detail["documentKeyID"] = e.DocumentKey.ID
	detail["operationType"] = e.OperationType
	detail["chainID"] = vaa.ChainID(e.DbFullDocument.ChainID).String()
	detail["emitterAddress"] = e.DbFullDocument.EmitterAddress
	detail["sequence"] = e.DbFullDocument.Sequence
	return detail
}
