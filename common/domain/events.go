package domain

import (
	"encoding/json"
	"time"
)

type NotificationEvent struct {
	TrackID string          `json:"trackId"`
	Source  string          `json:"source"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
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
	EmitterChain     int       `json:"emitterChain"`
	EmitterAddr      string    `json:"emitterAddr"`
	Sequence         string    `json:"sequence"`
	GuardianSetIndex int       `json:"guardianSetIndex"`
	Timestamp        time.Time `json:"timestamp"`
	Vaa              string    `json:"vaa"`
	TxHash           string    `json:"txHash"`
	Version          int       `json:"version"`
}

type PublishedLogMessage struct {
	ID           string    `json:"id"`
	EmitterChain int       `json:"emitterChain"`
	EmitterAddr  string    `json:"emitterAddr"`
	Sequence     string    `json:"sequence"`
	Timestamp    time.Time `json:"timestamp"`
	Vaa          string    `json:"vaa"`
	TxHash       string    `json:"txHash"`
}
