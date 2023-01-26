package docs

import (
	_ "embed"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type Controller struct {
	logger *zap.Logger
}

func NewController(logger *zap.Logger) *Controller {

	l := logger.With(zap.String("module", "DocsController"))

	return &Controller{logger: l}
}

//go:embed swagger.json
var swagger []byte

// GetSwagger godoc
// @Description Returns the swagger specification for this API.
// @Tags Wormscan
// @ID swagger
// @Success 200 {object} object
// @Failure 400
// @Failure 500
// @Router /swagger.json [get]
func (c *Controller) GetSwagger(ctx *fiber.Ctx) error {

	written, err := ctx.
		Response().
		BodyWriter().
		Write(swagger)

	if written != len(swagger) {
		return fmt.Errorf("partial write to response body: wrote %d bytes, expected %d", written, len(swagger))
	}

	return err
}
