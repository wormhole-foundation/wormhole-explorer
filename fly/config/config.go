package config

import (
	"fmt"
	"os"
	"strconv"
)

const defaultMaxHealthTimeSeconds = 60

// p2p network constants.
const (
	P2pMainNet = "mainnet"
	P2pTestNet = "testnet"
	P2pDevNet  = "devnet"
)

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
	case P2pMainNet:
		return &P2pNetworkConfig{P2pMainNet, MainNetP2ppNetworkID, MainNetP2pBootstrap, MainNetP2pPort}, nil
	case P2pTestNet:
		return &P2pNetworkConfig{P2pTestNet, TestNetP2ppNetworkID, TestNetP2pBootstrap, TestNetP2pPort}, nil
	case P2pDevNet:
		return &P2pNetworkConfig{P2pDevNet, DevNetP2ppNetworkID, DevNetP2pBootstrap, DevNetP2pPort}, nil
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
