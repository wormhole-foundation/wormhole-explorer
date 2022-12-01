package grpc

import (
	"testing"
	"time"

	spyv1 "github.com/certusone/wormhole/node/pkg/proto/spy/v1"
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
	var fi []filterSignedVaa
	svs := NewSignedVaaSubscribers()
	id, sub := svs.Register(fi)

	assert.NotEmpty(t, id)
	assert.NotNil(t, sub)
}

func TestSignedVaaSubscribers_Unregister(t *testing.T) {
	var fi []filterSignedVaa
	svs := NewSignedVaaSubscribers()
	id, _ := svs.Register(fi)

	assert.Equal(t, 1, len(svs.subscribers))
	svs.Unregister(id)
	assert.Equal(t, 0, len(svs.subscribers))
}

func TestSignedVaaSubscribers_HandleVAA(t *testing.T) {

	t.Run("empty filters", func(t *testing.T) {
		var fi []filterSignedVaa
		svs := NewSignedVaaSubscribers()
		_, sub := svs.Register(fi)

		vaas := []byte{0x0, 0x1, 0x2, 0x3}
		err := svs.HandleVAA(vaas)
		assert.Nil(t, err)
		msg := <-sub.ch
		assert.Equal(t, vaas, msg.vaaBytes)
	})

	t.Run("invalid vaa", func(t *testing.T) {
		fi := []filterSignedVaa{
			{
				chainId:     18,
				emitterAddr: vaa.Address{0x0, 0x1},
			},
		}
		svs := NewSignedVaaSubscribers()
		_, _ = svs.Register(fi)

		vaas := []byte{0x0, 0x1, 0x2, 0x3}
		err := svs.HandleVAA(vaas)
		assert.NotNil(t, err)
	})

	t.Run("filter doesn't apply", func(t *testing.T) {
		fi := []filterSignedVaa{
			{
				chainId:     18,
				emitterAddr: vaa.Address{0x0, 0x1},
			},
		}
		svs := NewSignedVaaSubscribers()
		_, sub := svs.Register(fi)
		vaa := createVAA(vaa.ChainIDEthereum, emitterAddr)
		vaaBytes, _ := vaa.MarshalBinary()
		err := svs.HandleVAA(vaaBytes)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(sub.ch))
	})
}

func TestAllVaaSubscribers_Register(t *testing.T) {
	var fi []*spyv1.FilterEntry
	logger := zaptest.NewLogger(t)
	avs := NewAllVaaSubscribers(logger)

	id, sub := avs.Register(fi)
	assert.NotEmpty(t, id)
	assert.NotNil(t, sub)
}

func TestAllVaaSubscribers_Unregister(t *testing.T) {
	var fi []*spyv1.FilterEntry
	logger := zaptest.NewLogger(t)
	avs := NewAllVaaSubscribers(logger)

	id, _ := avs.Register(fi)

	assert.Equal(t, 1, len(avs.subscribers))
	avs.Unregister(id)
	assert.Equal(t, 0, len(avs.subscribers))
}

func TestAllVaaSubscribers_HandleVAA(t *testing.T) {

	t.Run("empty filters", func(t *testing.T) {
		var fi []*spyv1.FilterEntry
		logger := zaptest.NewLogger(t)
		avs := NewAllVaaSubscribers(logger)
		_, sub := avs.Register(fi)

		emitterAddr := vaa.Address{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4}
		vaa := createVAA(vaa.ChainIDEthereum, emitterAddr)
		vaaBytes, _ := vaa.MarshalBinary()
		err := avs.HandleVAA(vaaBytes)
		assert.Nil(t, err)
		msg := <-sub.ch
		resp := msg.VaaType.(*spyv1.SubscribeSignedVAAByTypeResponse_SignedVaa)
		assert.Equal(t, vaaBytes, resp.SignedVaa.Vaa)
	})

	t.Run("invalid vaa", func(t *testing.T) {
		var fi []*spyv1.FilterEntry
		logger := zaptest.NewLogger(t)
		avs := NewAllVaaSubscribers(logger)
		_, _ = avs.Register(fi)

		vaas := []byte{0x0, 0x1, 0x2, 0x3}
		err := avs.HandleVAA(vaas)
		assert.NotNil(t, err)
	})

	t.Run("filter doesn't apply", func(t *testing.T) {
		fi := []*spyv1.FilterEntry{
			{
				Filter: &spyv1.FilterEntry_EmitterFilter{
					EmitterFilter: &spyv1.EmitterFilter{
						ChainId:        18,
						EmitterAddress: vaa.Address{0x0, 0x1}.String(),
					},
				},
			},
		}
		logger := zaptest.NewLogger(t)
		avs := NewAllVaaSubscribers(logger)
		_, sub := avs.Register(fi)
		emitterAddr := vaa.Address{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4}
		vaa := createVAA(vaa.ChainIDEthereum, emitterAddr)
		vaaBytes, _ := vaa.MarshalBinary()
		err := avs.HandleVAA(vaaBytes)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(sub.ch))
	})

}
