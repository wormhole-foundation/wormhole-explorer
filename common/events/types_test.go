package events

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetEventPayload contains a test harness for the `GetEventPayload` function.
func Test_GetEventPayload(t *testing.T) {

	body := `{
		"trackId": "63e16082da939a263512a307",
		"source": "fly",
		"event": "signed-vaa",
		"data": {
			"id": "2/000000000000000000000000f890982f9310df57d00f659cf4fd87e65aded8d7/162727",
			"emitterChain": 2,
			"emitterAddr": "000000000000000000000000f890982f9310df57d00f659cf4fd87e65aded8d7",
			"sequence": 162727,
			"guardianSetIndex": 0,
			"timestamp": "2023-08-04T11:43:48.000Z",
			"vaa": "010000000001005defe63f46c192b506758684fada6b97f5a8ee287a82efefa35c59dcf369a83b1abfe5431ad51a31051bf42851b5f699421e525745db03e8bc43a6b36dde6fc00064cd0ea4446900000002000000000000000000000000f890982f9310df57d00f659cf4fd87e65aded8d70000000000027ba7010300000000000000000000000000000000000000000000000000000000004c4b40000000000000000000000000b4fbf271143f4fbf7b91a5ded31805e42b2208d600026d9ae6b2d333c1d65301a59da3eed388ca5dc60cb12496584b75cbe6b15fdbed002000000000000000000000000072b916142650cb48bbbed0acaeb5b287d1c55d917b2262617369635f726563697069656e74223a7b22726563697069656e74223a22633256704d58426f4e445631626a646a4e6a426c6448566d6432317964575272617a4a3061336877647a4e6f595859794e6d4e6d5a6a5933227d7d",
			"txHash" : "406065c15b62426c51f987f5923fb376f6b60cb1c15724cc5460a08d18ccc337",
			"version" : 1
		}
	}`

	event := NotificationEvent{}
	err := json.Unmarshal([]byte(body), &event)
	assert.NoError(t, err)
	assert.Equal(t, "63e16082da939a263512a307", event.TrackID)
	assert.Equal(t, "fly", event.Source)
	assert.Equal(t, SignedVaaType, event.Event)
	signedVaa, err := GetEventData[SignedVaa](&event)
	assert.NoError(t, err)
	assert.Equal(t, "2/000000000000000000000000f890982f9310df57d00f659cf4fd87e65aded8d7/162727", signedVaa.ID)
}

func Test_GetEventPayload_Error(t *testing.T) {

	body := `{
		"trackId": "63e16082da939a263512a307",
		"source": "fly",
		"event": "signed-vaa"
	}`

	event := NotificationEvent{}
	err := json.Unmarshal([]byte(body), &event)
	assert.NoError(t, err)
	assert.Equal(t, "63e16082da939a263512a307", event.TrackID)
	assert.Equal(t, "fly", event.Source)
	assert.Equal(t, SignedVaaType, event.Event)
	_, err = GetEventData[SignedVaa](&event)
	assert.Error(t, err)
}
