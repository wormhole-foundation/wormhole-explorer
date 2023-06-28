package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
)

const defaultMaxHealthTimeSeconds = 60

// p2p network configuration constants.
const (
	// mainnet p2p config.
	MainNetP2ppNetworkID      = "/wormhole/mainnet/2"
	MainNetP2pBootstrap       = "/dns4/wormhole-mainnet-v2-bootstrap.certus.one/udp/8999/quic/p2p/12D3KooWQp644DK27fd3d4Km3jr7gHiuJJ5ZGmy8hH4py7fP4FP7"
	MainNetP2pPort       uint = 8999

	// testnet p2p config.
	TestNetP2ppNetworkID      = "/wormhole/testnet/2/1"
	TestNetP2pBootstrap       = "/dns4/wormhole-testnet-v2-bootstrap.certus.one/udp/8999/quic/p2p/12D3KooWAkB9ynDur1Jtoa97LBUp8RXdhzS5uHgAfdTquJbrbN7i"
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
	P2pPort      uint
}

// GetP2pNetwork get p2p network config.
func GetP2pNetwork() (*P2pNetworkConfig, error) {

	p2pEnviroment := os.Getenv("P2P_NETWORK")

	switch p2pEnviroment {
	case domain.P2pMainNet:
		return &P2pNetworkConfig{domain.P2pMainNet, MainNetP2ppNetworkID, MainNetP2pBootstrap, MainNetP2pPort}, nil
	case domain.P2pTestNet:
		return &P2pNetworkConfig{domain.P2pTestNet, TestNetP2ppNetworkID, TestNetP2pBootstrap, TestNetP2pPort}, nil
	case domain.P2pDevNet:
		return &P2pNetworkConfig{domain.P2pDevNet, DevNetP2ppNetworkID, DevNetP2pBootstrap, DevNetP2pPort}, nil
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

// GetEnviroment get enviroment.
func GetEnviroment() string {
	return os.Getenv("ENVIROMENT")
}

// GetAlertConfig get alert config.
func GetAlertConfig() (alert.AlertConfig, error) {
	p2pNetwork, err := GetP2pNetwork()
	if err != nil {
		return alert.AlertConfig{}, err
	}
	return alert.AlertConfig{
		Enviroment: GetEnviroment(),
		P2PNetwork: p2pNetwork.Enviroment,
		Enabled:    getAlertEnabled(),
		ApiKey:     getAlertApiKey(),
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

// GetMetricEnabled get if metric is enabled.
func GetMetricEnabled() bool {
	strMetricEnabled := os.Getenv("METRIC_ENABLED")
	metricEnabled, err := strconv.ParseBool(strMetricEnabled)
	if err != nil {
		metricEnabled = false
	}
	return metricEnabled
}

func GetPrefix() string {
	p2pNetwork, err := GetP2pNetwork()
	if err != nil {
		return ""
	}
	prefix := p2pNetwork.Enviroment + "-" + GetEnviroment() + ":"
	return prefix
}
