package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/config"
)

func ExtractDBLayer(ctx *fiber.Ctx) (string, error) {
	dbLayer := ctx.Get("DB_LAYER", config.DBLayerMongo)
	if dbLayer != config.DBLayerMongo && dbLayer != config.DBLayerPostgres {
		return "", fiber.NewError(fiber.StatusBadRequest, "invalid db layer")
	}
	return dbLayer, nil
}
