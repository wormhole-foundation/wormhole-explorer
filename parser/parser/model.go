package parser

import (
	"encoding/json"
	"math/big"
	"time"

	vaaPayloadParser "github.com/wormhole-foundation/wormhole-explorer/common/client/parser"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// ParsedVaaUpdate represent a parsed vaa update.
type ParsedVaaUpdate struct {
	ID                        string                                  `bson:"_id" json:"id"`
	EmitterChain              sdk.ChainID                             `bson:"emitterChain" json:"emitterChain"`
	EmitterAddr               string                                  `bson:"emitterAddr" json:"emitterAddr"`
	Sequence                  string                                  `bson:"sequence" json:"sequence"`
	AppIDs                    []string                                `bson:"appIds" json:"appIds"`
	ParsedPayload             interface{}                             `bson:"parsedPayload" json:"parsedPayload"`
	RawStandardizedProperties vaaPayloadParser.StandardizedProperties `bson:"rawStandardizedProperties" json:"rawStandardizedProperties"`
	StandardizedProperties    vaaPayloadParser.StandardizedProperties `bson:"standardizedProperties" json:"standardizedProperties"`
	UpdatedAt                 *time.Time                              `bson:"updatedAt" json:"updatedAt"`
	Timestamp                 time.Time                               `bson:"timestamp" json:"timestamp"`
}

// AttestationVaaProperties represent attestation vaa properties.
type AttestationVaaProperties struct {
	ID                string           `json:"id" db:"id"`
	VaaID             string           `json:"vaa_id" db:"vaa_id"`
	AppID             []string         `json:"app_id" db:"app_id"`
	Payload           *json.RawMessage `json:"payload" db:"payload"`
	PayloadType       *int             `json:"payload_type" db:"payload_type"`
	RawStandardFields *json.RawMessage `json:"raw_standard_fields" db:"raw_standard_fields"`
	FromChainID       *sdk.ChainID     `json:"from_chain_id" db:"from_chain_id"`
	FromAddress       *string          `json:"from_address" db:"from_address"`
	ToChainID         *sdk.ChainID     `json:"to_chain_id" db:"to_chain_id"`
	ToAddress         *string          `json:"to_address" db:"to_address"`
	TokenChainID      *sdk.ChainID     `json:"token_chain_id" db:"token_chain_id"`
	TokenAddress      *string          `json:"token_address" db:"token_address"`
	Amount            *big.Int         `json:"amount" db:"amount"`
	FeeChainID        *sdk.ChainID     `json:"fee_chain_id" db:"fee_chain_id"`
	FeeAddress        *string          `json:"fee_address" db:"fee_address"`
	Fee               *big.Int         `json:"fee" db:"fee"`
	Timestamp         time.Time        `json:"timestamp" db:"timestamp"`
	CreatedAt         time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at" db:"updated_at"`
}

type OperationAddress struct {
	ID          string    `db:"id"`
	Address     string    `db:"address"`
	AddressType string    `db:"address_type"`
	Timestamp   time.Time `db:"timestamp"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
