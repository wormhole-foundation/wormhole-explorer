package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/chains"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

func main() {

	// validate commandline arguments
	if len(os.Args) != 5 {
		log.Fatalf("Usage: ./%s <chain name> <tx hash> <block time> <p2p network>\n", os.Args[0])
	}

	// load rpc provider settings
	rpcProviderSettings, err := config.LoadFromEnv[config.RpcProviderSettings]()
	if err != nil {
		log.Fatalf("Failed to load credentials from environment: %v", err)
	}

	// load testnet rpc provider settings
	TestnetRpcProviderSettings, err := config.LoadFromEnv[config.TestnetRpcProviderSettings]()
	if err != nil {
		log.Fatalf("Failed to load credentials from environment: %v", err)
	}

	// get chain ID from args
	chainId, err := sdk.ChainIDFromString(os.Args[1])
	if err != nil {
		log.Fatalf("Failed to convert chain name to chain ID: %v", err)
	}

	blockTime, err := strconv.ParseInt(os.Args[3], 10, 64)
	if err != nil {
		log.Fatalf("Failed to convert block time to int64: %v", err)
	}
	timestamp := time.Unix(blockTime, 0)
	logger := logger.New("tx-tracker-fetch-one", logger.WithLevel("INFO"))

	// create rpc pool
	rpcPool, err := newRpcPool(*rpcProviderSettings, TestnetRpcProviderSettings)
	if err != nil {
		log.Fatalf("Failed to initialize rpc pool: %v", err)
	}

	// TODO: fix fetchOne
	txDetail, err := chains.FetchTx(context.Background(), rpcPool, chainId, os.Args[2], &timestamp, os.Args[4], logger)
	if err != nil {
		log.Fatalf("Failed to get transaction data: %v", err)
	}

	// print tx details
	log.Printf("tx detail: %+v", txDetail)
}

func newRpcPool(rpcSetting config.RpcProviderSettings,
	rpcTestnetSetting *config.TestnetRpcProviderSettings) (map[sdk.ChainID]*pool.Pool, error) {

	rpcPool := make(map[sdk.ChainID]*pool.Pool)

	// get rpc setings map
	rpcConfigMap, err := rpcSetting.ToMap()
	if err != nil {
		return nil, err
	}

	// get rpc testnet settings map
	var rpcTestnetMap map[sdk.ChainID][]config.RpcConfig
	if rpcTestnetSetting != nil {
		rpcTestnetMap, err = rpcTestnetSetting.ToMap()
		if err != nil {
			return nil, err
		}
	}

	// merge rpc testnet settings to rpc setting map
	if len(rpcTestnetMap) > 0 {
		for chainID, rpcConfig := range rpcTestnetMap {
			rpcConfigMap[chainID] = append(rpcConfigMap[chainID], rpcConfig...)
		}
	}

	// convert rpc settings map to rpc pool
	convertFn := func(rpcConfig []config.RpcConfig) []pool.Config {
		poolConfigs := make([]pool.Config, 0, len(rpcConfig))
		for _, rpc := range rpcConfig {
			poolConfigs = append(poolConfigs, pool.Config{
				Id:                rpc.Url,
				Priority:          rpc.Priority,
				RequestsPerMinute: rpc.RequestsPerMinute,
			})
		}
		return poolConfigs
	}

	// create rpc pool
	for chainID, rpcConfig := range rpcConfigMap {
		rpcPool[chainID] = pool.NewPool(convertFn(rpcConfig))
	}

	return rpcPool, nil
}
