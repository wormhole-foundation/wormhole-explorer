package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestHelloName(t *testing.T) {
	log := zap.NewExample()
	converter := NewNotificationEvent(log)
	msg := `{
		"trackId": "chain-event-5vbS8qQoaDGmUQmKhH32Y7g4VW63og3FqiqB2fy1HTaXuTmCymSi4XBmScBjXJnjQeMT38oLXe9ZVkuyjb4fLePf-0",
		"source": "blockchain-watcher",
		"event": "log-message-published",
		"timestamp": "2023-05-25T00:11:01Z",
		"version": "1",
		"data": {
		  "chainId": 1,
		  "emitterAddress": "ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5",
		  "txHash": "5vbS8qQoaDGmUQmKhH32Y7g4VW63og3FqiqB2fy1HTaXuTmCymSi4XBmScBjXJnjQeMT38oLXe9ZVkuyjb4fLePf",
		  "blockHeight": "227560241",
		  "blockTime": "2023-05-25T00:10:15-02:00",
		  "attributes": {
			"sender": "EVpwxYhvgXWxgosFhJv2UY5nFQXQ2GB143WPVqg7aKEj",
			"sequence": 321948,
			"nonce": 0,
			"consistencyLevel": 100
		  }
		}
	  }`
	event, err := converter(msg)
	assert.NoError(t, err)
	assert.NotNil(t, event)
	assert.Equal(t, "1/ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5/321948", event.ID)
	assert.Equal(t, uint16(1), event.ChainID)
	assert.Equal(t, "ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5", event.EmitterAddress)
}
