package events

import (
	"encoding/json"
	"time"
)

const (
	SignedVaaType           = "signed-vaa"
	LogMessagePublishedType = "log-message-published"
	EvmTransactionFoundType = "evm-transaction-found"
	TransferRedeemedType    = "transfer-redeemed"
	EvmTransferRedeemedName = "transfer-redeemed"
)

type NotificationEvent struct {
	TrackID   string          `json:"trackId"`
	Source    string          `json:"source"`
	Event     string          `json:"event"`
	Version   string          `json:"version"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

func NewNotificationEvent[T EventData](trackID, source, _type string, data T) (*NotificationEvent, error) {
	p, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return &NotificationEvent{
		TrackID:   trackID,
		Source:    source,
		Event:     _type,
		Data:      json.RawMessage(p),
		Version:   "1",
		Timestamp: time.Now(),
	}, nil
}

type EventData interface {
	SignedVaa | LogMessagePublished | EvmTransactionFound | TransferRedeemed
}

func GetEventData[T EventData](e *NotificationEvent) (T, error) {
	var data T
	err := json.Unmarshal(e.Data, &data)
	return data, err
}

type SignedVaa struct {
	ID               string    `json:"id"`
	EmitterChain     uint16    `json:"emitterChain"`
	EmitterAddress   string    `json:"emitterAddress"`
	Sequence         uint64    `json:"sequence"`
	GuardianSetIndex uint32    `json:"guardianSetIndex"`
	Timestamp        time.Time `json:"timestamp"`
	Vaa              []byte    `json:"vaa"`
	TxHash           string    `json:"txHash"`
	Version          int       `json:"version"`
}

type LogMessagePublished struct {
	ChainID     uint16                        `json:"chainId"`
	Emitter     string                        `json:"emitter"`
	TxHash      string                        `json:"txHash"`
	BlockHeight string                        `json:"blockHeight"`
	BlockTime   time.Time                     `json:"blockTime"`
	Attributes  LogMessagePublishedAttributes `json:"attributes"`
}

type LogMessagePublishedAttributes struct {
	Sender           string `json:"sender"`
	Sequence         uint64 `json:"sequence"`
	Nonce            uint32 `json:"nonce"`
	Payload          string `json:"payload"`
	ConsistencyLevel uint8  `json:"consistencyLevel"`
}

type EvmTransactionFound struct {
	ChainID     int                           `json:"chainId"`
	Emitter     string                        `json:"emitter"`
	TxHash      string                        `json:"txHash"`
	BlockHeight string                        `json:"blockHeight"`
	BlockTime   time.Time                     `json:"blockTime"`
	Attributes  EvmTransactionFoundAttributes `json:"attributes"`
}

type EvmTransactionFoundAttributes struct {
	Name           string `json:"name"`
	EmitterChain   int    `json:"emitterChain"`
	EmitterAddress string `json:"emitterAddress"`
	Sequence       uint64 `json:"sequence"`
	Method         string `json:"methodsByAddress"`
	From           string `json:"from"`
	To             string `json:"to"`
	Status         string `json:"status"`
}

type TransferRedeemed struct {
	ChainID     int                        `json:"chainId"`
	Emitter     string                     `json:"emitter"`
	TxHash      string                     `json:"txHash"`
	BlockHeight string                     `json:"blockHeight"`
	BlockTime   time.Time                  `json:"blockTime"`
	Attributes  TransferRedeemedAttributes `json:"attributes"`
}

type TransferRedeemedAttributes struct {
	EmitterChain   int    `json:"emitterChain"`
	EmitterAddress string `json:"emitterAddress"`
	Sequence       uint64 `json:"sequence"`
	Method         string `json:"methodsByAddress"`
	From           string `json:"from"`
	To             string `json:"to"`
	Status         string `json:"status"`
}
