package events

import (
	"encoding/json"
	"time"
)

const (
	SignedVaaType                 = "signed-vaa"
	LogMessagePublishedMesageType = "log-message-published"
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
	SignedVaa | LogMessagePublished
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
	Attributes  PublishedLogMessageAttributes `json:"attributes"`
}

type PublishedLogMessageAttributes struct {
	Sender           string `json:"sender"`
	Sequence         uint64 `json:"sequence"`
	Nonce            uint32 `json:"nonce"`
	Payload          string `json:"payload"`
	ConsistencyLevel uint8  `json:"consistencyLevel"`
}
