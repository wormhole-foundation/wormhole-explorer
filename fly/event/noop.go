package event

import (
	"context"
)

type NoopEventDispatcher struct{}

func NewNoopEventDispatcher() *NoopEventDispatcher {
	return &NoopEventDispatcher{}
}

func (n *NoopEventDispatcher) NewAttestationVaa(context.Context, Vaa) error {
	return nil
}

func (n *NoopEventDispatcher) NewDuplicateVaa(context.Context, DuplicateVaa) error {
	return nil
}

func (n *NoopEventDispatcher) NewGovernorStatus(context.Context, GovernorStatus) error {
	return nil
}

func (n *NoopEventDispatcher) NewGovernorConfig(ctx context.Context, e GovernorConfig) error {
	return nil
}
