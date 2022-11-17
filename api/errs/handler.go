// The errs package defines how errors are handled in the api.
// It define a type [AppError] that represent the api error response.
// Its define a custom error handling for the api.
package errs

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

// APIErrorHandler define a fiber custom error handler. This function process all errors
// returned from any handlers in the stack.
//
// To setup this function we must set the ErrorHandler field of the fiber.Config struct
// with this function and create a new fiber with this config.
//
// example: fiber.New(fiber.Config{ErrorHandler: errs.APIErrorHandler}
func APIErrorHandler(ctx *fiber.Ctx, err error) error {
	var apiError APIError
	switch {
	case errors.As(err, &apiError):
		ctx.Status(apiError.StatusCode).JSON(apiError)
	case errors.Is(err, ErrNotFound):
		apiError = NewNotFoundError()
		ctx.Status(fiber.StatusNotFound).JSON(apiError)
	default:
		apiError = NewInternalServerError()
		ctx.Status(fiber.StatusInternalServerError).JSON(apiError)
	}
	return nil
}
