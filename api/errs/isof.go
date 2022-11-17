// The errs package defines how errors are handled in the api.
// It define a type [AppError] that represent the api error response.
// Its define a custom error handling for the api.
package errs

import "errors"

// Error definitions to use in service/repository layers.
var (
	ErrInvalidParam  = errors.New("INVALID PARAM")
	ErrNotFound      = errors.New("NOT FOUND")
	ErrInternalError = errors.New("INTERNAL ERROR")
)

// IsOf reports whether any error in received chain matches target error list.
func IsOf(received error, targets ...error) bool {
	for _, t := range targets {
		if errors.Is(received, t) {
			return true
		}
	}
	return false
}
