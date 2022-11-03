package main

import (
	"context"
	"fmt"
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	ipfslog "github.com/ipfs/go-log/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/config"
	"github.com/wormhole-foundation/wormhole-explorer/api/db"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
	"github.com/wormhole-foundation/wormhole-explorer/api/observations"
	"github.com/wormhole-foundation/wormhole-explorer/api/vaa"
	"os"
	"strconv"
	"time"
)

var cacheConfig = cache.Config{
	Next: func(c *fiber.Ctx) bool {
		return c.Query("refresh") == "true"
	},
	Expiration:           1 * time.Second,
	CacheControl:         true,
	StoreResponseHeaders: true,
}

func healthOk(ctx *fiber.Ctx) error {
	ctx.Status(200)
	return ctx.SendString("Ok")
}

func main() {
	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Grab config
	cfg, err := config.Get()
	if err != nil {
		fmt.Fprint(os.Stderr, "Error parsing configuration")
		panic(err)
	}

	// Logging
	lvl, err := cfg.GetLogLevel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid logging level set: %v", cfg.LogLevel)
		panic(err)
	}

	rootLogger := ipfslog.Logger("wormhole-spy").Desugar()
	ipfslog.SetAllLoggers(lvl)

	// Setup DB
	cli, err := db.Connect(appCtx, cfg.DB.URL)
	if err != nil {
		panic(err)
	}
	db := cli.Database(cfg.DB.Name)

	// Setup repositories
	vaaRepo := vaa.NewRepository(db, rootLogger)
	obsRepo := observations.NewRepository(db, rootLogger)

	// Setup services
	vaaService := vaa.NewService(vaaRepo, rootLogger)
	obsService := observations.NewService(obsRepo, rootLogger)

	// Setup controllers
	vaaCtrl := vaa.NewController(vaaService, rootLogger)
	observationsCtrl := observations.NewController(obsService, rootLogger)

	// Setup API
	app := fiber.New()

	// Middleware
	prometheus := fiberprometheus.New("wormscan")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format: "level=info timestamp=${time} method=${method} path=${path} status${status} request_id=${locals:requestid}\n",
	}))

	api := app.Group("/api")
	api.Use(cors.New()) // TODO CORS restrictions?
	api.Use(middleware.ExtractPagination)

	api.Get("/health", healthOk)

	vaas := api.Group("/vaas")
	vaas.Use(cache.New(cacheConfig))
	vaas.Get("/", vaaCtrl.FindAll)
	vaas.Get("/:chain", vaaCtrl.FindByChain)
	vaas.Get("/:chain/:emitter", vaaCtrl.FindByEmitter)
	vaas.Get("/:chain/:emitter/:sequence", vaaCtrl.FindById)
	api.Get("vaa-counts", vaaCtrl.GetStats)
	api.Get("vaas-sans-pythnet", vaaCtrl.FindForPythnet)

	observations := api.Group("/observations")
	observations.Get("/", observationsCtrl.FindAll)
	observations.Get("/:chain", observationsCtrl.FindAllByChain)
	observations.Get("/:chain/:emitter", observationsCtrl.FindAllByEmitter)
	observations.Get("/:chain/:emitter/:sequence", observationsCtrl.FindAllByVAA)
	observations.Get("/:chain/:emitter/:sequence/:signer/:hash", observationsCtrl.FindOne)

	app.Listen(":" + strconv.Itoa(cfg.PORT))
}
