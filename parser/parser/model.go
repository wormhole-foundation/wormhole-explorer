package parser

import (
	"time"

	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// StandardizedProperties represent a standardized properties.
type StandardizedProperties struct {
	AppIds       []string    `json:"appIds" bson:"appIds"`
	FromChain    sdk.ChainID `json:"fromChain" bson:"fromChain"`
	FromAddress  string      `json:"fromAddress" bson:"fromAddress"`
	ToChain      sdk.ChainID `json:"toChain" bson:"toChain"`
	ToAddress    string      `json:"toAddress" bson:"toAddress"`
	TokenChain   sdk.ChainID `json:"tokenChain" bson:"tokenChain"`
	TokenAddress string      `json:"tokenAddress" bson:"tokenAddress"`
	Amount       string      `json:"amount" bson:"amount"`
	FeeAddress   string      `json:"feeAddress" bson:"feeAddress"`
	FeeChain     sdk.ChainID `json:"feeChain" bson:"feeChain"`
	Fee          string      `json:"fee" bson:"fee"`
}

// ParsedVaaUpdate represent a parsed vaa update.
type ParsedVaaUpdate struct {
	ID                        string                 `bson:"_id" json:"id"`
	EmitterChain              sdk.ChainID            `bson:"emitterChain" json:"emitterChain"`
	EmitterAddr               string                 `bson:"emitterAddr" json:"emitterAddr"`
	Sequence                  string                 `bson:"sequence" json:"sequence"`
	AppIDs                    []string               `bson:"appIds" json:"appIds"`
	ParsedPayload             interface{}            `bson:"parsedPayload" json:"parsedPayload"`
	RawStandardizedProperties StandardizedProperties `bson:"rawStandardizedProperties" json:"rawStandardizedProperties"`
	StandardizedProperties    StandardizedProperties `bson:"standardizedProperties" json:"standardizedProperties"`
	UpdatedAt                 *time.Time             `bson:"updatedAt" json:"updatedAt"`
	Timestamp                 time.Time              `bson:"-" json:"-"`
}
