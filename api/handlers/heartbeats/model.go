package heartbeats

import "time"

// HeartbeatDoc represent an heartbeat document.
type HeartbeatDoc struct {
	ID            string             `bson:"_id" json:"id"`
	BootTimestamp int64              `bson:"boottimestamp" json:"bootTimestamp"`
	Counter       int64              `bson:"counter" json:"counter"`
	Features      []string           `bson:"features" json:"features"`
	GuardianAddr  string             `bson:"guardianaddr" json:"guardianAddr"`
	IndexedAt     *time.Time         `bson:"indexedAt" json:"indexedAt"`
	NodeName      string             `bson:"nodename" json:"nodeName"`
	Timestamp     int64              `bson:"timestamp" json:"timestamp"`
	UpdatedAt     *time.Time         `bson:"updatedAt" json:"updatedAt"`
	Version       string             `bson:"version" json:"version"`
	Networks      []HeartbeatNetwork `bson:"networks" json:"networks"`
}

// HeartbeatNetwork definition.
type HeartbeatNetwork struct {
	ID              int64  `bson:"id" json:"id"`
	Height          int64  `bson:"height" json:"height"`
	ContractAddress string `bson:"contractaddress" json:"contractAddress"`
	ErrorCount      int64  `bson:"errorcount" json:"errorCount"`
}
