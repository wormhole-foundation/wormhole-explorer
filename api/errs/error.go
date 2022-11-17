// The errs package defines how errors are handled in the api.
// It define a type [AppError] that represent the api error response.
// Its define a custom error handling for the api.
package errs

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// API error codes. These error code are the same used in guardian API.
const (
	InvalidParam = 3
	NotFound     = 5
	Internal     = 13
)

// APIError api error response.
// This structure is defined to be aligned with the way the guardian API handles the error response.
type APIError struct {
	StatusCode int      `json:"-"`
	Code       int      `json:"code"`
	Message    string   `json:"message"`
	Details    []string `json:"details"`
}

// NewApiError create a new APIError.
func NewApiError(statusCode, code int, message string) APIError {
	return APIError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Details:    []string{},
	}
}

// NewInternalServerError create a new APIError for Internal Server Errors.
func NewInternalServerError() APIError {
	return APIError{
		StatusCode: fiber.StatusInternalServerError,
		Code:       Internal,
		Message:    ErrInternalError.Error(),
		Details:    []string{},
	}
}

// NewNotFoundError create a new APIError for Not Found errors.
func NewNotFoundError() APIError {
	return APIError{
		StatusCode: fiber.StatusNotFound,
		Code:       NotFound,
		Message:    ErrNotFound.Error(),
		Details:    []string{},
	}
}

// NewParamError create a new APIError for invalid param errors.
func NewParamError(message string) APIError {
	if message == "" {
		message = ErrInvalidParam.Error()
	}
	return APIError{
		StatusCode: fiber.StatusBadRequest,
		Code:       InvalidParam,
		Message:    message,
		Details:    []string{},
	}
}

// NewHTTPErrorWithDetails create a new APIError with details.
func NewHTTPErrorWithDetails(statusCode, code int, message string, details []string) APIError {
	return APIError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Details:    details,
	}
}

// Error interface implementation.
func (h APIError) Error() string {
	return fmt.Sprintf("code: %d, message: %s, details: %v", h.Code, h.Message, h.Details)
}
