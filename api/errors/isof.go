package errors

import "errors"

func IsOf(received error, targets ...error) bool {
	for _, t := range targets {
		if errors.Is(received, t) {
			return true
		}
	}
	return false
}
