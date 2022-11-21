package errors

import "errors"

// Error definitions to use in service and repository layers.
var (
	ErrMalformedQuery = errors.New("MALFORMED_QUERY")
	ErrNotFound       = errors.New("NOT FOUND")
	ErrInternalError  = errors.New("INTERNAL ERROR")
)
