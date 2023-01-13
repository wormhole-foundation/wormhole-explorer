package parser

import (
	"time"
)

// ParsedVaaUpdate representa a parsedVaa document.
type ParsedVaaUpdate struct {
	ID           string      `bson:"_id"`
	EmitterChain uint16      `bson:"emitterChain"`
	EmitterAddr  string      `bson:"emitterAddr"`
	Sequence     string      `bson:"sequence"`
	Result       interface{} `bson:"result"`
	Timestamp    *time.Time  `bson:"timestamp"`
	UpdatedAt    *time.Time  `bson:"updatedAt"`
}
