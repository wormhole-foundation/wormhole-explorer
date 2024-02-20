package activity

import (
	"context"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/wormhole-foundation/wormhole-explorer/jobs/jobs/protocols/activity/internal/repositories"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// Protocols
const (
	MayanProtocol     = "mayan"
	AllBridgeProtocol = "allbridge"
)

type ProtocolsActivityJob struct {
	statsDB          api.WriteAPIBlocking
	logger           *zap.Logger
	activityFetchers []ClientActivity
	version          string
}

// ClientActivity Abstraction for fetching protocol Activity since each client may have different implementation details.
type ClientActivity interface {
	Get(ctx context.Context, from, to time.Time) (repositories.ProtocolActivity[repositories.Activity], error)
	ProtocolName() string
}

// ActivitiesClientsFactory RestClient Factory to create the right client for each protocol.
var ActivitiesClientsFactory = map[string]func(name, url string, logger *zap.Logger) ClientActivity{

	MayanProtocol: func(name, url string, logger *zap.Logger) ClientActivity {
		return repositories.NewMayanRestClient(name, url, logger, &http.Client{})
	},

	AllBridgeProtocol: func(name, url string, logger *zap.Logger) ClientActivity {
		return repositories.NewAllBridgeRestClient(name, url, logger, &http.Client{})
	},
}
