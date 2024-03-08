package config

import (
	"context"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
)

// p2p network configuration constants.
const (
	// mainnet p2p config.
	MainNetP2ppNetworkID      = "/wormhole/mainnet/2"
	MainNetP2pBootstrap       = "/dns4/wormhole-v2-mainnet-bootstrap.xlabs.xyz/udp/8999/quic-v1/p2p/12D3KooWNQ9tVrcb64tw6bNs2CaNrUGPM7yRrKvBBheQ5yCyPHKC,/dns4/wormhole.mcf.rocks/udp/8999/quic-v1/p2p/12D3KooWDZVv7BhZ8yFLkarNdaSWaB43D6UbQwExJ8nnGAEmfHcU,/dns4/wormhole-v2-mainnet-bootstrap.staking.fund/udp/8999/quic-v1/p2p/12D3KooWG8obDX9DNi1KUwZNu9xkGwfKqTp2GFwuuHpWZ3nQruS1"
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

type Configuration struct {
	P2pNetwork                string `env:"P2P_NETWORK,required"`
	Environment               string `env:"ENVIRONMENT,required"`
	LogLevel                  string `env:"LOG_LEVEL,default=warn"`
	MongoUri                  string `env:"MONGODB_URI,required"`
	MongoDatabase             string `env:"MONGODB_DATABASE,required"`
	MongoEnableQueryLog       bool   `env:"MONGODB_ENABLE_QUERY_LOG"`
	ObservationsChannelSize   int    `env:"OBSERVATIONS_CHANNEL_SIZE,required"`
	VaasChannelSize           int    `env:"VAAS_CHANNEL_SIZE,required"`
	HeartbeatsChannelSize     int    `env:"HEARTBEATS_CHANNEL_SIZE,required"`
	GovernorConfigChannelSize int    `env:"GOVERNOR_CONFIG_CHANNEL_SIZE,required"`
	GovernorStatusChannelSize int    `env:"GOVERNOR_STATUS_CHANNEL_SIZE,required"`
	VaasWorkersSize           int    `env:"VAAS_WORKERS_SIZE,default=5"`
	ObservationsWorkersSize   int    `env:"OBSERVATIONS_WORKERS_SIZE,default=10"`
	AlertEnabled              bool   `env:"ALERT_ENABLED"`
	AlertApiKey               string `env:"ALERT_API_KEY"`
	MetricsEnabled            bool   `env:"METRICS_ENABLED"`
	ApiPort                   uint   `env:"API_PORT,required"`
	P2pPort                   uint   `env:"P2P_PORT,required"`
	PprofEnabled              bool   `env:"PPROF_ENABLED"`
	MaxHealthTimeSeconds      int64  `env:"MAX_HEALTH_TIME_SECONDS,default=60"`
	IsLocal                   bool
	Redis                     *RedisConfiguration
	Aws                       *AwsConfiguration
	ObservationsDedup         Cache `env:", prefix=OBSERVATIONS_DEDUP_,required"`
	ObservationsTxHash        Cache `env:", prefix=OBSERVATIONS_TX_HASH_,required"`
	VaasDedup                 Cache `env:", prefix=VAAS_DEDUP_,required"`
}

type RedisConfiguration struct {
	RedisUri        string `env:"REDIS_URI,required"`
	RedisPrefix     string `env:"REDIS_PREFIX,required"`
	RedisVaaChannel string `env:"REDIS_VAA_CHANNEL,required"`
}

type AwsConfiguration struct {
	AwsRegion          string `env:"AWS_REGION,required"`
	AwsAccessKeyID     string `env:"AWS_ACCESS_KEY_ID"`
	AwsSecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY"`
	AwsEndpoint        string `env:"AWS_ENDPOINT"`
	SqsUrl             string `env:"SQS_URL,required"`
	ObservationsSqsUrl string `env:"OBSERVATIONS_SQS_URL,required"`
}

type Cache struct {
	ExpirationInSeconds int64 `env:"CACHE_EXPIRATION_SECONDS,required"`
	NumKeys             int64 `env:"CACHE_NUM_KEYS,required"`
	MaxCostsInMB        int64 `env:"CACHE_MAX_COSTS_MB,required"`
}

// New creates a configuration with the values from .env file and environment variables.
func New(ctx context.Context, isLocal *bool) (*Configuration, error) {
	_ = godotenv.Load(".env", "../.env")

	var configuration Configuration
	if err := envconfig.Process(ctx, &configuration); err != nil {
		return nil, err
	}

	configuration.IsLocal = isLocal != nil && *isLocal

	if !configuration.IsLocal {
		var redis RedisConfiguration
		if err := envconfig.Process(ctx, &redis); err != nil {
			return nil, err
		}
		configuration.Redis = &redis

		var aws AwsConfiguration
		if err := envconfig.Process(ctx, &aws); err != nil {
			return nil, err
		}
		configuration.Aws = &aws
	}

	return &configuration, nil
}

// GetP2pNetwork get p2p network config.
func (c *Configuration) GetP2pNetwork() (*P2pNetworkConfig, error) {

	p2pEnviroment := c.P2pNetwork

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

func (c *Configuration) GetPrefix() string {
	prefix := c.P2pNetwork + "-" + c.Environment
	return prefix
}
