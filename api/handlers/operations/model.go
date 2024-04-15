package operations

import (
	"time"

	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// OperationDto operation data transfer object.
type OperationDto struct {
	ID                     string                  `bson:"_id"`
	TxHash                 string                  `bson:"txHash"`
	Symbol                 string                  `bson:"symbol"`
	UsdAmount              string                  `bson:"usdAmount"`
	TokenAmount            string                  `bson:"tokenAmount"`
	Vaa                    *VaaDto                 `bson:"vaa"`
	SourceTx               *OriginTx               `bson:"originTx" json:"originTx"`
	DestinationTx          *DestinationTx          `bson:"destinationTx" json:"destinationTx"`
	Payload                map[string]any          `bson:"payload"`
	StandardizedProperties *StandardizedProperties `bson:"standardizedProperties"`
}

// StandardizedProperties represents the standardized properties of a operation.
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

// VaaDto vaa data transfer object.
type VaaDto struct {
	ID                string      `bson:"_id" json:"id"`
	Version           uint8       `bson:"version" json:"version"`
	EmitterChain      sdk.ChainID `bson:"emitterChain" json:"emitterChain"`
	EmitterAddr       string      `bson:"emitterAddr" json:"emitterAddr"`
	EmitterNativeAddr string      `json:"emitterNativeAddr,omitempty"`
	Sequence          string      `bson:"sequence" json:"-"`
	GuardianSetIndex  uint32      `bson:"guardianSetIndex" json:"guardianSetIndex"`
	Vaa               []byte      `bson:"vaas" json:"vaa"`
	Timestamp         *time.Time  `bson:"timestamp" json:"timestamp"`
	UpdatedAt         *time.Time  `bson:"updatedAt" json:"updatedAt"`
	IndexedAt         *time.Time  `bson:"indexedAt" json:"indexedAt"`
	Hash              []byte      `bson:"hash" json:"hash"`
	IsDuplicated      bool        `bson:"isDuplicated" json:"isDuplicated"`
}

// GlobalTransactionDoc definitions.
type GlobalTransactionDoc struct {
	ID            string         `bson:"_id" json:"id"`
	OriginTx      *OriginTx      `bson:"originTx" json:"originTx"`
	DestinationTx *DestinationTx `bson:"destinationTx" json:"destinationTx"`
}

// OriginTx represents a origin transaction.
type OriginTx struct {
	TxHash    string        `bson:"nativeTxHash" json:"txHash"`
	From      string        `bson:"from" json:"from"`
	Status    string        `bson:"status" json:"status"`
	Timestamp *time.Time    `bson:"timestamp" json:"timestamp"`
	Attribute *AttributeDoc `bson:"attribute" json:"attribute"`
}

// AttributeDoc represents a custom attribute for a origin transaction.
type AttributeDoc struct {
	Type  string         `bson:"type" json:"type"`
	Value map[string]any `bson:"value" json:"value"`
}

// DestinationTx represents a destination transaction.
type DestinationTx struct {
	ChainID     sdk.ChainID `bson:"chainId" json:"chainId"`
	Status      string      `bson:"status" json:"status"`
	Method      string      `bson:"method" json:"method"`
	TxHash      string      `bson:"txHash" json:"txHash"`
	From        string      `bson:"from" json:"from"`
	To          string      `bson:"to" json:"to"`
	BlockNumber string      `bson:"blockNumber" json:"blockNumber"`
	Timestamp   *time.Time  `bson:"timestamp" json:"timestamp"`
	UpdatedAt   *time.Time  `bson:"updatedAt" json:"updatedAt"`
}
