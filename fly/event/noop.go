package event

import "context"

type NoopEventDispatcher struct{}

func NewNoopEventDispatcher() *NoopEventDispatcher {
	return &NoopEventDispatcher{}
}

func (n *NoopEventDispatcher) NewDuplicateVaa(context.Context, DuplicateVaa) error {
	return nil
}

func (n *NoopEventDispatcher) NewGovernorStatus(context.Context, GovernorStatus) error {
	return nil
}
