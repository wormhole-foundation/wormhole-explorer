package main

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/improbable-eng/grpc-web/go/grpcweb"

	ipfslog "github.com/ipfs/go-log/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/governor"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/heartbeats"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/infrastructure"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/observations"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/vaa"
	wormscanCache "github.com/wormhole-foundation/wormhole-explorer/api/internal/cache"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/config"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/db"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
	"github.com/wormhole-foundation/wormhole-explorer/api/response"
	"github.com/wormhole-foundation/wormhole-explorer/api/routes/guardian"
	"github.com/wormhole-foundation/wormhole-explorer/api/routes/wormscan"
	rpcApi "github.com/wormhole-foundation/wormhole-explorer/api/rpc"
	"go.uber.org/zap"
)

//go:embed docs/swagger.json
var swagger []byte

// GetSwagger godoc
// @Description Returns the swagger specification for this API.
// @Tags Wormscan
// @ID swagger
// @Success 200 {object} object
// @Failure 400
// @Failure 500
// @Router /swagger.json [get]
func GetSwagger(ctx *fiber.Ctx) error {

	written, err := ctx.
		Response().
		BodyWriter().
		Write(swagger)

	if written != len(swagger) {
		return fmt.Errorf("partial write to response body: wrote %d bytes, expected %d", written, len(swagger))
	}

	return err
}

// @title Wormhole Guardian API
// @version 1.0
// @description Wormhole Guardian API
// @description To get information from the Wormhole Network.
// @description Check each endpoint documentation for more information.
// @termsOfService https://wormhole.com/
// @contact.name API Support
// @contact.url http://wormhole.com/support
// @contact.email info@wormhole.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath /v1
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

	rootLogger := ipfslog.Logger("wormhole-api").Desugar()
	ipfslog.SetAllLoggers(lvl)

	// Setup DB
	cli, err := db.Connect(appCtx, cfg.DB.URL)
	if err != nil {
		panic(err)
	}
	db := cli.Database(cfg.DB.Name)

	// Get cache get function
	cacheGetFunc := NewCache(cfg, rootLogger)

	// Set up repositories
	vaaRepo := vaa.NewRepository(db, rootLogger)
	obsRepo := observations.NewRepository(db, rootLogger)
	governorRepo := governor.NewRepository(db, rootLogger)
	infrastructureRepo := infrastructure.NewRepository(db, rootLogger)
	heartbeatsRepo := heartbeats.NewRepository(db, rootLogger)

	// Set up services
	vaaService := vaa.NewService(vaaRepo, cacheGetFunc, rootLogger)
	obsService := observations.NewService(obsRepo, rootLogger)
	governorService := governor.NewService(governorRepo, rootLogger)
	infrastructureService := infrastructure.NewService(infrastructureRepo, rootLogger)
	heartbeatsService := heartbeats.NewService(heartbeatsRepo, rootLogger)

	// Set up a custom error handler
	response.SetEnableStackTrace(*cfg)
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})

	// Configure middleware
	prometheus := fiberprometheus.New("wormscan")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)
	app.Use(cors.New())
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format: "level=info timestamp=${time} method=${method} path=${path} status${status} request_id=${locals:requestid}\n",
	}))

	// Set up route handlers
	app.Get("/swagger.json", GetSwagger)
	wormscan.RegisterRoutes(app, rootLogger, vaaService, obsService, governorService, infrastructureService)
	guardian.RegisterRoutes(app, rootLogger, vaaService, governorService, heartbeatsService)

	// Set up gRPC handlers
	handler := rpcApi.NewHandler(vaaService, heartbeatsService, governorService, rootLogger)
	grpcServer := rpcApi.NewServer(handler, rootLogger)
	grpcWebServer := grpcweb.WrapServer(grpcServer)
	app.Use(
		adaptor.HTTPMiddleware(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if grpcWebServer.IsGrpcWebRequest(r) {
					grpcWebServer.ServeHTTP(w, r)
				} else {
					next.ServeHTTP(w, r)
				}
			})
		}))

	rootLogger.Fatal("http listen", zap.Error(app.Listen(":"+strconv.Itoa(cfg.PORT))))
}

// NewCache return a CacheGetFunc to get a value by a Key from cache.
func NewCache(cfg *config.AppConfig, looger *zap.Logger) wormscanCache.CacheGetFunc {
	if cfg.RunMode == config.RunModeDevelopmernt && !cfg.Cache.Enabled {
		dummyCacheClient := wormscanCache.NewDummyCacheClient()
		return dummyCacheClient.Get
	}
	cacheClient := wormscanCache.NewCacheClient(cfg.Cache.URL, cfg.Cache.Enabled, looger)
	return cacheClient.Get
}
