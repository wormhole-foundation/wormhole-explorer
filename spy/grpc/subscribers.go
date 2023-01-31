package grpc

import (
	"context"
	"fmt"

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
	id      string
	filters []filterSignedVaa
	ch      chan message
}
type subscriptionAllVaa struct {
	id      string
	filters []*spyv1.FilterEntry
	ch      chan *spyv1.SubscribeSignedVAAByTypeResponse
}

func subscriptionId() string {
	return uuid.New().String()
}

// SignedVaaSubscribers represents signed VAA subscribers.
type SignedVaaSubscribers struct {
	source           chan []byte
	subscribers      map[string]*subscriptionSignedVaa
	addSubscriber    chan *subscriptionSignedVaa
	removeSubscriber chan *subscriptionSignedVaa
	logger           *zap.Logger
}

// NewSignedVaaSubscribers creates a signed VAA subscribers.
func NewSignedVaaSubscribers(logger *zap.Logger) *SignedVaaSubscribers {
	return &SignedVaaSubscribers{
		subscribers:      make(map[string]*subscriptionSignedVaa),
		addSubscriber:    make(chan *subscriptionSignedVaa, 1),
		removeSubscriber: make(chan *subscriptionSignedVaa, 1),
		source:           make(chan []byte, 1),
		logger:           logger,
	}
}

// AllVaaSubscribers represents all VAA subscribers.
type AllVaaSubscribers struct {
	source           chan []byte
	subscribers      map[string]*subscriptionAllVaa
	addSubscriber    chan *subscriptionAllVaa
	removeSubscriber chan *subscriptionAllVaa
	logger           *zap.Logger
}

// NewAllVaaSubscribers creates all VAA subscribers.
func NewAllVaaSubscribers(logger *zap.Logger) *AllVaaSubscribers {
	return &AllVaaSubscribers{
		subscribers:      make(map[string]*subscriptionAllVaa),
		addSubscriber:    make(chan *subscriptionAllVaa, 1),
		removeSubscriber: make(chan *subscriptionAllVaa, 1),
		source:           make(chan []byte, 1),
		logger:           logger,
	}
}

// Register registers a new subscriber with a list of filters.
func (s *SignedVaaSubscribers) Register(fi []filterSignedVaa) *subscriptionSignedVaa {
	sub := &subscriptionSignedVaa{
		id:      subscriptionId(),
		ch:      make(chan message, 1),
		filters: fi,
	}
	s.logger.Info("Registering subscriber in signed VAAs ...", zap.String("id", sub.id))
	s.addSubscriber <- sub
	return sub
}

// Unregister removes a subscriber.
func (s *SignedVaaSubscribers) Unregister(sub *subscriptionSignedVaa) {
	s.logger.Info("Unregistering subscriber in signed VAAs ...", zap.String("id", sub.id))
	s.removeSubscriber <- sub
}

// HandleVAA sends a VAA to subscribers that filters apply the conditions.
func (s *SignedVaaSubscribers) HandleVAA(vaas []byte) error {
	s.source <- vaas
	return nil
}

func (s *SignedVaaSubscribers) Start(ctx context.Context) {
	defer func() {
		for _, subscriberByID := range s.subscribers {
			if subscriberByID != nil {
				close(subscriberByID.ch)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case newSubscriber := <-s.addSubscriber:
			s.subscribers[newSubscriber.id] = newSubscriber
			s.logger.Info("New subscriber registered in signed VAAs", zap.String("id", newSubscriber.id))
		case subscriberToRemove := <-s.removeSubscriber:
			if subscriber, exists := s.subscribers[subscriberToRemove.id]; exists {
				close(subscriber.ch)
				delete(s.subscribers, subscriberToRemove.id)
				s.logger.Info("Subscriber unregistered in signed VAAs", zap.String("id", subscriber.id))
			}
		case vaas, ok := <-s.source:
			if !ok {
				break
			}
			var v *vaa.VAA

			for _, sub := range s.subscribers {
				if len(sub.filters) == 0 {
					select {
					case sub.ch <- message{vaaBytes: vaas}:
					default:
					}
					continue
				}

				if v == nil {
					var err error
					v, err = vaa.Unmarshal(vaas)
					if err != nil {
						s.logger.Error("Unmarshal vaa in signed VAAs", zap.Error(err))
						break
					}
				}

				for _, fi := range sub.filters {
					if fi.chainId == v.EmitterChain && fi.emitterAddr == v.EmitterAddress {
						select {
						case sub.ch <- message{vaaBytes: vaas}:
						default:
						}
					}
				}

			}
		}
	}
}

// Register registers a new subscriber with a list of filters.
func (s *AllVaaSubscribers) Register(fi []*spyv1.FilterEntry) *subscriptionAllVaa {
	sub := &subscriptionAllVaa{
		id:      subscriptionId(),
		ch:      make(chan *spyv1.SubscribeSignedVAAByTypeResponse, 1),
		filters: fi,
	}
	s.logger.Info("Registering subscriber in all VAAs ...", zap.String("id", sub.id))
	s.addSubscriber <- sub
	return sub
}

// Unregister removes a subscriber.
func (s *AllVaaSubscribers) Unregister(sub *subscriptionAllVaa) {
	s.logger.Info("Unregistering subscriber in all VAAs ...", zap.String("id", sub.id))
	s.removeSubscriber <- sub
}

// HandleVAA sends a VAA to subscribers that filters apply the conditions.
func (s *AllVaaSubscribers) HandleVAA(vaaBytes []byte) error {
	s.source <- vaaBytes
	return nil
}

func (s *AllVaaSubscribers) Start(ctx context.Context) {
	defer func() {
		for _, subscriberByID := range s.subscribers {
			if subscriberByID != nil {
				close(subscriberByID.ch)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case newSubscriber := <-s.addSubscriber:
			s.subscribers[newSubscriber.id] = newSubscriber
			s.logger.Info("New subscriber registered in all VAAs", zap.String("id", newSubscriber.id))
		case subscriberToRemove := <-s.removeSubscriber:
			if subscriber, exists := s.subscribers[subscriberToRemove.id]; exists {
				close(subscriber.ch)
				delete(s.subscribers, subscriberToRemove.id)
				s.logger.Info("Subscriber unregistered in all VAAs", zap.String("id", subscriber.id))
			}
		case vaaBytes, ok := <-s.source:
			if !ok {
				break
			}
			v, err := vaa.Unmarshal(vaaBytes)
			if err != nil {
				s.logger.Error("failed unmarshaing VAA bytes from gossipv1.SignedVAAWithQuorum.", zap.Error(err))
				continue
			}

			// resType defines which oneof proto will be retuned - res type "SignedVaa" is *gossipv1.SignedVAAWithQuorum
			resType := &spyv1.SubscribeSignedVAAByTypeResponse_SignedVaa{
				SignedVaa: &gossipv1.SignedVAAWithQuorum{Vaa: vaaBytes},
			}

			// envelope is the highest level proto struct, the wrapper proto that contains one of the VAA types.
			envelope := &spyv1.SubscribeSignedVAAByTypeResponse{
				VaaType: resType,
			}

			// loop through the subscriptions and send responses to everyone that wants this VAA
			for _, sub := range s.subscribers {
				if len(sub.filters) == 0 {
					// this subscription has no filters, send them the VAA.
					select {
					case sub.ch <- envelope:
					default:
					}
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
							select {
							case sub.ch <- envelope:
							default:
							}
						}
					default:
						s.logger.Error(fmt.Sprintf("Unsupported filter type in subscriptions: %T", filter))
					}
				}

			}
		}
	}
}
