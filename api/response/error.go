// The response package defines the success and error response type.
// It define a type [AppError] that represent the api error response.
// Its define a custom error handling for the api.
package response

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/config"
)

// API error codes. These error code are the same used in guardian API.
const (
	InvalidParam = 3
	NotFound     = 5
	Internal     = 13
)

var enableStackTrace bool

// SetEnableStackTrace enable/disable send the stacktrace field in the response.
func SetEnableStackTrace(cfg config.AppConfig) {
	if cfg.RunMode == config.RunModeDevelopmernt {
		enableStackTrace = true
		return
	}
	enableStackTrace = false
}

// APIError api error response.
// This structure is defined to be aligned with the way the guardian API handles the error response.
type APIError struct {
	StatusCode int           `json:"-"`
	Code       int           `json:"code"` // support to guardian-api code.
	Message    string        `json:"message"`
	Details    []ErrorDetail `json:"details"`
}

// ErrorDetail definition.
// This structure contains the requestID and the stacktrace of the error.
type ErrorDetail struct {
	RequestID  string `json:"request_id"`
	StackTrace string `json:"stack_trace,omitempty"`
}

// Error interface implementation.
func (a APIError) Error() string {
	return fmt.Sprintf("code: %d, message: %s, details: %v", a.Code, a.Message, a.Details)
}

// NewApiError create a new api response.
func NewApiError(ctx *fiber.Ctx, statusCode, code int, message string, err error) APIError {
	detail := ErrorDetail{
		RequestID: fmt.Sprintf("%v", ctx.Locals("requestid")),
	}
	if enableStackTrace && err != nil {
		detail.StackTrace = fmt.Sprintf("%+v\n", err)
	}
	return APIError{
		StatusCode: fiber.StatusBadRequest,
		Code:       InvalidParam,
		Message:    message,
		Details:    []ErrorDetail{detail},
	}
}

// NewInvalidParamError create a invalid param Error.
func NewInvalidParamError(ctx *fiber.Ctx, message string, err error) APIError {
	if message == "" {
		message = "INVALID PARAM"
	}
	detail := ErrorDetail{
		RequestID: fmt.Sprintf("%v", ctx.Locals("requestid")),
	}
	if enableStackTrace && err != nil {
		detail.StackTrace = fmt.Sprintf("%+v\n", err)
	}
	return APIError{
		StatusCode: fiber.StatusBadRequest,
		Code:       InvalidParam,
		Message:    message,
		Details:    []ErrorDetail{detail},
	}
}

// NewInternalError create a new APIError for Internal Errors.
func NewInternalError(ctx *fiber.Ctx, err error) APIError {
	detail := ErrorDetail{
		RequestID: fmt.Sprintf("%v", ctx.Locals("requestid")),
	}
	if enableStackTrace && err != nil {
		detail.StackTrace = fmt.Sprintf("%+v\n", err)
	}
	return APIError{
		StatusCode: fiber.StatusInternalServerError,
		Code:       Internal,
		Message:    "INTERNAL ERROR",
		Details:    []ErrorDetail{detail},
	}
}

// NewNotFoundError create a new APIError for Not Found errors.
func NewNotFoundError(ctx *fiber.Ctx) APIError {
	return APIError{
		StatusCode: fiber.StatusNotFound,
		Code:       NotFound,
		Message:    "NOT FOUND",
		Details: []ErrorDetail{{
			RequestID: fmt.Sprintf("%v", ctx.Locals("requestid")),
		}},
	}
}
