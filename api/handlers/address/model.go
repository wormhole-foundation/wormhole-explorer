package address

import "github.com/wormhole-foundation/wormhole-explorer/api/handlers/vaa"

type AddressOverview struct {
	Vaas []*vaa.VaaDoc `json:"vaas"`
}

//// VaaDoc defines the JSON model for VAA objects in the REST API.
//type VaaDoc struct {
//	ID               string      `bson:"_id" json:"id"`
//	Version          uint8       `bson:"version" json:"version"`
//	EmitterChain     vaa.ChainID `bson:"emitterChain" json:"emitterChain"`
//	EmitterAddr      string      `bson:"emitterAddr" json:"emitterAddr"`
//	Sequence         string      `bson:"sequence" json:"-"`
//	GuardianSetIndex uint32      `bson:"guardianSetIndex" json:"guardianSetIndex"`
//	Vaa              []byte      `bson:"vaas" json:"vaa"`
//	Timestamp        *time.Time  `bson:"timestamp" json:"timestamp"`
//	UpdatedAt        *time.Time  `bson:"updatedAt" json:"updatedAt"`
//	IndexedAt        *time.Time  `bson:"indexedAt" json:"indexedAt"`
//	// TxHash is an extension field - it is not present in the guardian API.
//	TxHash *string `bson:"txHash" json:"txHash"`
//	// AppId is an extension field - it is not present in the guardian API.
//	AppId string `bson:"appId" json:"appId,omitempty"`
//	// Payload is an extension field - it is not present in the guardian API.
//	Payload map[string]interface{} `bson:"payload" json:"payload,omitempty"`
//	// NativeTxHash is an extension field - it is not present in the guardian API.
//	NativeTxHash string `bson:"nativeTxHash" json:"nativeTxHash,omitempty"`
//}
//
//// MarshalJSON interface implementation.
//func (v *VaaDoc) MarshalJSON() ([]byte, error) {
//	sequence, err := strconv.ParseUint(v.Sequence, 10, 64)
//	if err != nil {
//		return []byte{}, err
//	}
//
//	type Alias VaaDoc
//	return json.Marshal(&struct {
//		Sequence uint64 `json:"sequence"`
//		*Alias
//	}{
//		Sequence: sequence,
//		Alias:    (*Alias)(v),
//	})
//}
//
//// VaaStats definition.
//type VaaStats struct {
//	ChainID vaa.ChainID `bson:"_id" json:"chainId"`
//	Count   int64       `bson:"count" json:"count"`
//}
//
