package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/config"
)

func extractDBLayer(ctx *fiber.Ctx) string {
	return ctx.Get("DB_LAYER", config.DBLayerMongo)
}
