package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSourceChainAttributes(t *testing.T) {
	e := &Event{
		TrackID:        "pipeline-1",
		Type:           SourceChainEvent,
		ID:             "2/0000000000000000000000003ee18b2214aff97000d974cf647e7c347e8fa585/107429",
		ChainID:        2,
		EmitterAddress: "0000000000000000000000003ee18b2214aff97000d974cf647e7c347e8fa585",
		Sequence:       "107429",
		Timestamp:      nil,
		TxHash:         "0x7837b71e9d83b4fbff385ed4af3f70e13b820c2ba6580494bc1a205f3cd8e88c",
		Attributes:     &SourceChainAttributes{},
	}
	_, ok := GetAttributes[*SourceChainAttributes](e)
	assert.True(t, ok)
}

func TestGetTargetChainAttributes(t *testing.T) {
	e := &Event{
		TrackID:        "chain-event-1",
		Type:           TargetChainEvent,
		ID:             "2/0000000000000000000000003ee18b2214aff97000d974cf647e7c347e8fa585/107429",
		ChainID:        2,
		EmitterAddress: "9NQ5FxpSHLKBt4tHZwqp5zXKq2yFMTqtGR5iQX1BH1Ay",
		Sequence:       "377892",
		Timestamp:      nil,
		TxHash:         "YQ3XZQ33Uu2TV78Ms6zPtznxK5aWK3zJAbmawi46rtb126cNpnJ9B3CfK5EjTJUoYKkJp8QbTRiEsBkxD8nzDD9",
		Attributes: &TargetChainAttributes{
			Emitter:     "9NQ5FxpSHLKBt4tHZwqp5zXKq2yFMTqtGR5iQX1BH1Ay",
			BlockHeight: "183675392",
		},
	}
	attr, ok := GetAttributes[*TargetChainAttributes](e)
	assert.True(t, ok)
	assert.NotNil(t, attr)
	assert.Equal(t, "9NQ5FxpSHLKBt4tHZwqp5zXKq2yFMTqtGR5iQX1BH1Ay", attr.Emitter)
}
