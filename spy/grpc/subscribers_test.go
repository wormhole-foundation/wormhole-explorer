package grpc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap/zaptest"
)

var emitterAddr = vaa.Address{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4}

func createVAA(chainID vaa.ChainID, emitterAddr vaa.Address) *vaa.VAA {
	var payload = []byte{97, 97, 97, 97, 97, 97}

	vaa := &vaa.VAA{
		Version:          vaa.SupportedVAAVersion,
		GuardianSetIndex: uint32(1),
		Signatures:       nil,
		Timestamp:        time.Unix(0, 0),
		Nonce:            uint32(1),
		Sequence:         uint64(1),
		ConsistencyLevel: uint8(32),
		EmitterChain:     chainID,
		EmitterAddress:   emitterAddr,
		Payload:          payload,
	}

	return vaa
}

func TestSignedVaaSubscribers_Register(t *testing.T) {
	logger := zaptest.NewLogger(t)
	var fi []filterSignedVaa
	svs := NewSignedVaaSubscribers(logger)
	sub := svs.Register(fi)
	assert.NotNil(t, sub)
	assert.NotEmpty(t, sub.id)
}

func TestSignedVaaSubscribers_Unregister(t *testing.T) {
	logger := zaptest.NewLogger(t)
	var fi []filterSignedVaa
	svs := NewSignedVaaSubscribers(logger)
	sub := svs.Register(fi)
	assert.Equal(t, 1, len(svs.addSubscriber))
	svs.Unregister(sub)
	assert.Equal(t, 1, len(svs.removeSubscriber))
}

func TestSignedVaaSubscribers_HandleVAA(t *testing.T) {

	t.Run("empty filters", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		var fi []filterSignedVaa
		svs := NewSignedVaaSubscribers(logger)
		svs.Register(fi)

		vaas := []byte{0x0, 0x1, 0x2, 0x3}
		err := svs.HandleVAA(vaas)
		assert.Nil(t, err)
		vaaBytes := <-svs.source
		assert.Equal(t, vaas, vaaBytes)
	})

	t.Run("invalid vaa", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		fi := []filterSignedVaa{
			{
				chainId:     18,
				emitterAddr: vaa.Address{0x0, 0x1},
			},
		}
		svs := NewSignedVaaSubscribers(logger)
		_ = svs.Register(fi)

		vaas := []byte{0x0, 0x1, 0x2, 0x3}
		err := svs.HandleVAA(vaas)
		assert.Nil(t, err)
	})

	t.Run("filter doesn't apply", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		fi := []filterSignedVaa{
			{
				chainId:     18,
				emitterAddr: vaa.Address{0x0, 0x1},
			},
		}
		svs := NewSignedVaaSubscribers(logger)
		sub := svs.Register(fi)
		vaa := createVAA(vaa.ChainIDEthereum, emitterAddr)
		vaaBytes, _ := vaa.MarshalBinary()
		err := svs.HandleVAA(vaaBytes)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(sub.ch))
	})
}
