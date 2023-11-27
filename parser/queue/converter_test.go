package queue

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

func TestNotificationEvent(t *testing.T) {
	log := zap.NewExample()
	converter := NewNotificationEvent(log)
	msg := `{
		"trackId":"chain-event-0xb437df51c6c9af58eff89e328f84d8bece25d718baf088899c9692782fe29c43-10012515",
		"source":"blockchain-watcher",
		"event":"log-message-published",
		"timestamp":"2023-11-10T14:20:45.159Z",
		"version":"1",
		"data":{
			"chainId":2,
			"emitter":"0x706abc4e45d419950511e474c7b9ed348a4a716c",
			"txHash":"0xb437df51c6c9af58eff89e328f84d8bece25d718baf088899c9692782fe29c43",
			"blockHeight":"10012515",
			"blockTime":"2023-11-09T09:06:24.000Z",
			"attributes":{
				"sender":"0xe9d87dD072B0bcE6aA9335d590cfB0342870d7B0",
				"sequence":1,
				"payload":"0x7b226e65766572223a7b22676f6e6e61223a7b2267697665223a7b22796f75223a227570227d7d7d7d",
				"nonce":1699520760,
				"consistencyLevel":200
			}
		}
	}`
	event, err := converter(msg)
	assert.NoError(t, err)
	assert.NotNil(t, event)
	assert.Equal(t, "chain-event-0xb437df51c6c9af58eff89e328f84d8bece25d718baf088899c9692782fe29c43-10012515", event.TrackID)
	assert.Equal(t, "2/000000000000000000000000e9d87dd072b0bce6aa9335d590cfb0342870d7b0/1", event.ID)
	assert.Equal(t, uint16(2), event.ChainID)
	assert.Equal(t, "0xe9d87dD072B0bcE6aA9335d590cfB0342870d7B0", event.EmitterAddress)
	vaa, err := sdk.Unmarshal(event.Vaa)
	assert.NoError(t, err)
	assert.NotNil(t, vaa)
	expectedPayload := []byte{123, 34, 110, 101, 118, 101, 114, 34, 58, 123, 34, 103, 111, 110, 110, 97, 34, 58, 123, 34, 103, 105, 118, 101, 34, 58, 123, 34, 121, 111, 117, 34, 58, 34, 117, 112, 34, 125, 125, 125, 125}
	assert.Equal(t, expectedPayload, vaa.Payload)
}

func TestSqsEvent(t *testing.T) {
	log := zap.NewExample()
	converter := NewNotificationEvent(log)
	msg := `
	{
		"Type" : "Notification",
		"MessageId" : "14d855ca-ad78-59c5-b30e-0802e1362944",
		"SequenceNumber" : "10000000000040190002",
		"TopicArn" : "arn:aws:sns:us-east-2:581679387567:notification-chain-events-dev-testnet.fifo",
		"Subject" : "blockchain-watcher",
		"Message" : "{\"trackId\":\"chain-event-0xb6b7af602aa098fbd8c88da2c2e4a316eef22f0ee621c5ca7616992c3fd9d3fe-10012893\",\"source\":\"blockchain-watcher\",\"event\":\"log-message-published\",\"timestamp\":\"2023-11-10T15:19:42.320Z\",\"version\":\"1\",\"data\":{\"chainId\":2,\"emitter\":\"0x706abc4e45d419950511e474c7b9ed348a4a716c\",\"txHash\":\"0xb6b7af602aa098fbd8c88da2c2e4a316eef22f0ee621c5ca7616992c3fd9d3fe\",\"blockHeight\":\"10012893\",\"blockTime\":\"2023-11-09T10:41:24.000Z\",\"attributes\":{\"sender\":\"0x28D8F1Be96f97C1387e94A53e00eCcFb4E75175a\",\"sequence\":3418,\"payload\":\"0x010017000000000000000000000000b5b6bf4224f75762dae40c862dc899431ea1778300000040000000000000000000000000000000000000000000000000000000000000000300000000000000000000000000000000000000000000000000000000654cb74d00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000007a12000000000000000000000000000000000000000000000000000000006a6fd66db0002000000000000000000000000e310c47fa3e3f011a6e3108e3c725cff4900199b00000000000000000000000090995dbd1aae85872451b50a569de947d34ac4ee000000000000000000000000d1463b4fe86166768d2ff51b1a928bebb5c9f375000000000000000000000000e310c47fa3e3f011a6e3108e3c725cff4900199b00\",\"nonce\":0,\"consistencyLevel\":200}}}",
		"Timestamp" : "2023-11-10T15:19:42.548Z",
		"UnsubscribeURL" : "https://sns.us-east-2.amazonaws.com/?Action=Unsubscribe&SubscriptionArn=arn:aws:sns:us-east-2:581679387567:notification-chain-events-dev-testnet.fifo:2e1cc196-afd8-4efb-b9b3-27c38e688494"
	  }
	`

	// unmarshal body to sqsEvent from sns/sqs subscription
	var sqsEvent sqsEvent
	err := json.Unmarshal([]byte(msg), &sqsEvent)
	assert.NoError(t, err)
	event, err := converter(sqsEvent.Message)
	assert.NoError(t, err)
	assert.NotNil(t, event)
	assert.Equal(t, "2/00000000000000000000000028d8f1be96f97c1387e94a53e00eccfb4e75175a/3418", event.ID)
	assert.Equal(t, uint16(2), event.ChainID)
	assert.Equal(t, "0x28D8F1Be96f97C1387e94A53e00eCcFb4E75175a", event.EmitterAddress)
	assert.Equal(t, "0xb6b7af602aa098fbd8c88da2c2e4a316eef22f0ee621c5ca7616992c3fd9d3fe", event.TxHash)
	vaa, err := sdk.Unmarshal(event.Vaa)
	assert.NoError(t, err)
	assert.NotNil(t, vaa)
	expectedTimestamp, err := time.Parse(time.RFC3339, "2023-11-09T10:41:24.000Z")
	assert.NoError(t, err)
	assert.Equal(t, uint64(3418), vaa.Sequence)
	assert.Equal(t, sdk.ChainIDEthereum, vaa.EmitterChain)
	assert.Equal(t, expectedTimestamp.UTC(), vaa.Timestamp.UTC())
	assert.Equal(t, uint8(200), vaa.ConsistencyLevel)
}
