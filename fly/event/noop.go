package event

import (
	"context"

	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type NoopEventDispatcher struct{}

func NewNoopEventDispatcher() *NoopEventDispatcher {
	return &NoopEventDispatcher{}
}

func (n *NoopEventDispatcher) NewVaa(context.Context, sdk.VAA) error {
	return nil
}

func (n *NoopEventDispatcher) NewDuplicateVaa(context.Context, DuplicateVaa) error {
	return nil
}

func (n *NoopEventDispatcher) NewGovernorStatus(context.Context, GovernorStatus) error {
	return nil
}
