package domain

import (
	"encoding/json"
	"time"
)

const (
	SignedVaaType           = "signed-vaa"
	PublishedLogMessageType = "published-log-message"
)

type NotificationEvent struct {
	TrackID string          `json:"trackId"`
	Source  string          `json:"source"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func NewNotificationEvent[T EventPayload](trackID, source, _type string, payload T) (*NotificationEvent, error) {
	p, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return &NotificationEvent{
		TrackID: trackID,
		Source:  source,
		Type:    _type,
		Payload: json.RawMessage(p),
	}, nil
}

type EventPayload interface {
	SignedVaa | PublishedLogMessage
}

func GetEventPayload[T EventPayload](e *NotificationEvent) (T, error) {
	var payload T
	err := json.Unmarshal(e.Payload, &payload)
	return payload, err
}

type SignedVaa struct {
	ID               string    `json:"id"`
	EmitterChain     uint16    `json:"emitterChain"`
	EmitterAddr      string    `json:"emitterAddr"`
	Sequence         uint64    `json:"sequence"`
	GuardianSetIndex uint32    `json:"guardianSetIndex"`
	Timestamp        time.Time `json:"timestamp"`
	Vaa              []byte    `json:"vaa"`
	TxHash           string    `json:"txHash"`
	Version          int       `json:"version"`
}

type PublishedLogMessage struct {
	ID           string    `json:"id"`
	EmitterChain uint16    `json:"emitterChain"`
	EmitterAddr  string    `json:"emitterAddr"`
	Sequence     uint64    `json:"sequence"`
	Timestamp    time.Time `json:"timestamp"`
	Vaa          []byte    `json:"vaa"`
	TxHash       string    `json:"txHash"`
}
