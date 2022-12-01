package grpc

import (
	"fmt"
	"sync"

	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
	spyv1 "github.com/certusone/wormhole/node/pkg/proto/spy/v1"
	"github.com/google/uuid"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type message struct {
	vaaBytes []byte
}

type filterSignedVaa struct {
	chainId     vaa.ChainID
	emitterAddr vaa.Address
}
type subscriptionSignedVaa struct {
	filters []filterSignedVaa
	ch      chan message
}
type subscriptionAllVaa struct {
	filters []*spyv1.FilterEntry
	ch      chan *spyv1.SubscribeSignedVAAByTypeResponse
}

func subscriptionId() string {
	return uuid.New().String()
}

// SignedVaaSubscribers represents signed VAA subscribers.
type SignedVaaSubscribers struct {
	m           sync.Mutex
	subscribers map[string]*subscriptionSignedVaa
}

// NewSignedVaaSubscribers creates a signed VAA subscribers.
func NewSignedVaaSubscribers() *SignedVaaSubscribers {
	return &SignedVaaSubscribers{subscribers: make(map[string]*subscriptionSignedVaa)}
}

// AllVaaSubscribers represents all VAA subscribers.
type AllVaaSubscribers struct {
	m           sync.Mutex
	subscribers map[string]*subscriptionAllVaa
	logger      *zap.Logger
}

// NewAllVaaSubscribers creates all VAA subscribers.
func NewAllVaaSubscribers(logger *zap.Logger) *AllVaaSubscribers {
	return &AllVaaSubscribers{subscribers: make(map[string]*subscriptionAllVaa), logger: logger}
}

// Register registers a new subscriber with a list of filters.
func (s *SignedVaaSubscribers) Register(fi []filterSignedVaa) (string, *subscriptionSignedVaa) {
	s.m.Lock()
	id := subscriptionId()
	sub := &subscriptionSignedVaa{
		ch:      make(chan message, 1),
		filters: fi,
	}
	s.subscribers[id] = sub
	s.m.Unlock()
	return id, sub
}

// Unregister removes a subscriber.
func (s *SignedVaaSubscribers) Unregister(id string) {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.subscribers, id)
}

// HandleVAA sends a VAA to subscribers that filters apply the conditions.
func (s *SignedVaaSubscribers) HandleVAA(vaas []byte) error {
	s.m.Lock()
	defer s.m.Unlock()

	var v *vaa.VAA

	for _, sub := range s.subscribers {
		if len(sub.filters) == 0 {
			sub.ch <- message{vaaBytes: vaas}
			continue
		}

		if v == nil {
			var err error
			v, err = vaa.Unmarshal(vaas)
			if err != nil {
				return err
			}
		}

		for _, fi := range sub.filters {
			if fi.chainId == v.EmitterChain && fi.emitterAddr == v.EmitterAddress {
				sub.ch <- message{vaaBytes: vaas}
			}
		}

	}
	return nil
}

// Register registers a new subscriber with a list of filters.
func (s *AllVaaSubscribers) Register(fi []*spyv1.FilterEntry) (string, *subscriptionAllVaa) {
	s.m.Lock()
	id := subscriptionId()
	sub := &subscriptionAllVaa{
		ch:      make(chan *spyv1.SubscribeSignedVAAByTypeResponse, 1),
		filters: fi,
	}
	s.subscribers[id] = sub
	s.m.Unlock()
	return id, sub
}

// Unregister removes a subscriber.
func (s *AllVaaSubscribers) Unregister(id string) {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.subscribers, id)
}

// HandleVAA sends a VAA to subscribers that filters apply the conditions.
func (s *AllVaaSubscribers) HandleVAA(vaaBytes []byte) error {

	v, err := vaa.Unmarshal(vaaBytes)
	if err != nil {
		s.logger.Error("failed unmarshaing VAA bytes from gossipv1.SignedVAAWithQuorum.", zap.Error(err))
		return err
	}

	// resType defines which oneof proto will be retuned - res type "SignedVaa" is *gossipv1.SignedVAAWithQuorum
	resType := &spyv1.SubscribeSignedVAAByTypeResponse_SignedVaa{
		SignedVaa: &gossipv1.SignedVAAWithQuorum{Vaa: vaaBytes},
	}

	// envelope is the highest level proto struct, the wrapper proto that contains one of the VAA types.
	envelope := &spyv1.SubscribeSignedVAAByTypeResponse{
		VaaType: resType,
	}

	s.m.Lock()
	defer s.m.Unlock()

	// loop through the subscriptions and send responses to everyone that wants this VAA
	for _, sub := range s.subscribers {
		if len(sub.filters) == 0 {
			// this subscription has no filters, send them the VAA.
			sub.ch <- envelope
			continue
		}

		// this subscription has filters.
		for _, filterEntry := range sub.filters {
			filter := filterEntry.GetFilter()
			switch t := filter.(type) {
			case *spyv1.FilterEntry_EmitterFilter:
				filterAddr := t.EmitterFilter.EmitterAddress
				filterChain := vaa.ChainID(t.EmitterFilter.ChainId)

				if v.EmitterChain == filterChain && v.EmitterAddress.String() == filterAddr {
					// it is a match, send the response
					sub.ch <- envelope
				}
			default:
				s.logger.Error(fmt.Sprintf("unsupported filter type in subscriptions: %T", filter))
			}
		}

	}

	return nil
}
