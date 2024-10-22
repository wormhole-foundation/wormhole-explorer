package governor_status

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/domain"
)

// Set generic type definition.
type Set[T comparable] map[T]struct{}

// add a value to the set
func (s Set[T]) Add(v T) {
	s[v] = struct{}{}
}

// check if the set contains a value
func (s Set[T]) Contains(v T) bool {
	_, ok := s[v]
	return ok
}

// get length of the set
func (s Set[T]) Len() int {
	return len(s)
}

// to slice
func (s Set[T]) ToSlice() []T {
	var slice []T
	for k := range s {
		slice = append(slice, k)
	}
	return slice
}

type Params struct {
	TrackID         string
	NodeGovernorVaa *domain.NodeGovernorVaa
}

// ProcessorFunc is a function to process a governor message.
type ProcessorFunc func(context.Context, *Params) error
