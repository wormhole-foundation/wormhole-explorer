package parser

import (
	"time"
)

// VaaParserFunctions represent a vaaParserFunctions document.
type VaaParserFunctions struct {
	ID             string     `bson:"_id"`
	CreatedAt      *time.Time `bson:"createdAt"`
	UpdatedAt      *time.Time `bson:"updatedAt"`
	EmitterChain   uint16     `bson:"emitterChain"`
	EmitterAddress string     `bson:"emitterAddress"`
	ParserFunction string     `bson:"parserFunction"`
}

// ParsedVaaUpdate representa a parsedVaa document.
type ParsedVaaUpdate struct {
	ID           string     `bson:"_id"`
	EmitterChain uint16     `bson:"emitterChain"`
	EmitterAddr  string     `bson:"emitterAddr"`
	Sequence     string     `bson:"sequence"`
	Result       string     `bson:"result"`
	Timestamp    *time.Time `bson:"timestamp"`
	UpdatedAt    *time.Time `bson:"updatedAt"`
}
