package config

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
)

const defaultMaxHealthTimeSeconds = 60

// p2p network configuration constants.
const (
	// mainnet p2p config.
	MainNetP2ppNetworkID      = "/wormhole/mainnet/2"
	MainNetP2pBootstrap       = "/dns4/wormhole-v2-mainnet-bootstrap.xlabs.xyz/udp/8999/quic/p2p/12D3KooWNQ9tVrcb64tw6bNs2CaNrUGPM7yRrKvBBheQ5yCyPHKC,/dns4/wormhole.mcf.rocks/udp/8999/quic/p2p/12D3KooWDZVv7BhZ8yFLkarNdaSWaB43D6UbQwExJ8nnGAEmfHcU,/dns4/wormhole-v2-mainnet-bootstrap.staking.fund/udp/8999/quic/p2p/12D3KooWG8obDX9DNi1KUwZNu9xkGwfKqTp2GFwuuHpWZ3nQruS1"
	MainNetP2pPort       uint = 8999

	// testnet p2p config.
	TestNetP2ppNetworkID      = "/wormhole/testnet/2/1"
	TestNetP2pBootstrap       = "/dns4/wormhole-testnet-v2-bootstrap.certus.one/udp/8999/quic-v1/p2p/12D3KooWAkB9ynDur1Jtoa97LBUp8RXdhzS5uHgAfdTquJbrbN7i,/dns4/t-guardian-01.nodes.stable.io/udp/8999/quic-v1/p2p/12D3KooWCW3LGUtkCVkHZmVSZHzL3C4WRKWfqAiJPz1NR7dT9Bxh,/dns4/t-guardian-02.nodes.stable.io/udp/8999/quic-v1/p2p/12D3KooWJXA6goBCiWM8ucjzc4jVUBSqL9Rri6UpjHbkMPErz5zK"
	TestNetP2pPort       uint = 8999

	// devnet p2p config.
	DevNetP2ppNetworkID      = "/wormhole/dev"
	DevNetP2pBootstrap       = "/dns4/guardian-0.guardian/udp/8999/quic/p2p/12D3KooWL3XJ9EMCyZvmmGXL2LMiVBtrVa2BuESsJiXkSj7333Jw"
	DevNetP2pPort       uint = 8999
)

// P2pNetworkConfig config struct.
type P2pNetworkConfig struct {
	Enviroment   string
	P2pNetworkID string
	P2pBootstrap string
}

// GetP2pNetwork get p2p network config.
func GetP2pNetwork() (*P2pNetworkConfig, error) {

	p2pEnviroment := os.Getenv("P2P_NETWORK")

	switch p2pEnviroment {
	case domain.P2pMainNet:
		return &P2pNetworkConfig{domain.P2pMainNet, MainNetP2ppNetworkID, MainNetP2pBootstrap}, nil
	case domain.P2pTestNet:
		return &P2pNetworkConfig{domain.P2pTestNet, TestNetP2ppNetworkID, TestNetP2pBootstrap}, nil
	case domain.P2pDevNet:
		return &P2pNetworkConfig{domain.P2pDevNet, DevNetP2ppNetworkID, DevNetP2pBootstrap}, nil
	default:
		return nil, fmt.Errorf(`invalid P2P_NETWORK enviroment variable: "%s"`, p2pEnviroment)
	}
}

// GetPprofEnabled get if pprof is enabled.
func GetPprofEnabled() bool {
	strPprofEnable := os.Getenv("PPROF_ENABLED")
	pprofEnabled, _ := strconv.ParseBool(strPprofEnable)
	return pprofEnabled
}

// GetMaxHealthTimeSeconds get MaxHealthTimeSeconds env value.
func GetMaxHealthTimeSeconds() int64 {
	var maxHealthTimeSeconds int
	strMaxHealthTimeSeconds := os.Getenv("MAX_HEALTH_TIME_SECONDS")
	maxHealthTimeSeconds, err := strconv.Atoi(strMaxHealthTimeSeconds)
	if err != nil {
		maxHealthTimeSeconds = defaultMaxHealthTimeSeconds
	}
	return int64(maxHealthTimeSeconds)
}

// GetEnvironment get environment.
func GetEnvironment() string {
	return os.Getenv("ENVIRONMENT")
}

// GetAlertConfig get alert config.
func GetAlertConfig() (alert.AlertConfig, error) {
	return alert.AlertConfig{
		Environment: GetEnvironment(),
		Enabled:     getAlertEnabled(),
		ApiKey:      getAlertApiKey(),
	}, nil
}

// getAlertEnabled get if alert is enabled.
func getAlertEnabled() bool {
	strAlertEnabled := os.Getenv("ALERT_ENABLED")
	alertEnabled, err := strconv.ParseBool(strAlertEnabled)
	if err != nil {
		alertEnabled = false
	}
	return alertEnabled
}

// getAlertApiKey get alert api key.
func getAlertApiKey() string {
	return os.Getenv("ALERT_API_KEY")
}

// GetMetricsEnabled get if metrics is enabled.
func GetMetricsEnabled() bool {
	strMetricsEnabled := os.Getenv("METRICS_ENABLED")
	metricsEnabled, err := strconv.ParseBool(strMetricsEnabled)
	if err != nil {
		metricsEnabled = false
	}
	return metricsEnabled
}

func GetPrefix() string {
	p2pNetwork, err := GetP2pNetwork()
	if err != nil {
		return ""
	}
	prefix := p2pNetwork.Enviroment + "-" + GetEnvironment()
	return prefix
}

type Configuration struct {
	ObservationsChannelSize   int  `env:"OBSERVATIONS_CHANNEL_SIZE,required"`
	VaasChannelSize           int  `env:"VAAS_CHANNEL_SIZE,required"`
	HeartbeatsChannelSize     int  `env:"HEARTBEATS_CHANNEL_SIZE,required"`
	GovernorConfigChannelSize int  `env:"GOVERNOR_CONFIG_CHANNEL_SIZE,required"`
	GovernorStatusChannelSize int  `env:"GOVERNOR_STATUS_CHANNEL_SIZE,required"`
	ObservationsWorkersSize   int  `env:"OBSERVATIONS_WORKERS_SIZE,default=10"`
	ApiPort                   uint `env:"API_PORT,required"`
	P2pPort                   uint `env:"P2P_PORT,required"`
}

// New creates a configuration with the values from .env file and environment variables.
func New(ctx context.Context) (*Configuration, error) {
	_ = godotenv.Load(".env", "../.env")

	var configuration Configuration
	if err := envconfig.Process(ctx, &configuration); err != nil {
		return nil, err
	}

	return &configuration, nil
}
