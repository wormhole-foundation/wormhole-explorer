package grpc

import (
	"context"

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
