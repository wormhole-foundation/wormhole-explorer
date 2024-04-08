package main

import (
	"github.com/spf13/cobra"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/cmd/backfiller"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/cmd/service"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

func main() {
	_ = execute()
}

func execute() error {
	root := &cobra.Command{
		Use: "tx-tracker",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				service.Run()
			}
		},
	}

	addServiceCommand(root)
	addBackfiller(root)

	return root.Execute()
}

func addServiceCommand(root *cobra.Command) {
	serviceCommand := &cobra.Command{
		Use:   "service",
		Short: "Run tx-tracker as service",
		Run: func(_ *cobra.Command, _ []string) {
			service.Run()
		},
	}
	root.AddCommand(serviceCommand)
}

func addBackfiller(parent *cobra.Command) {
	backfiller := &cobra.Command{
		Use: "backfiller",
	}

	addBackfillerByVaas(backfiller)
	parent.AddCommand(backfiller)
}

func addBackfillerByVaas(parent *cobra.Command) {
	var mongoUri, mongoDb, logLevel, startTime, endTime, p2pNetwork, emitterAddress, rpcProvidersPath string
	var numWorkers int
	var emitterChainID uint16
	var pageSize, requestsPerMinute int64
	var overwrite, disableDBUpsert bool

	vaas := &cobra.Command{
		Use:   "vaas",
		Short: "Run backfiller for vaas",
		Run: func(_ *cobra.Command, _ []string) {
			cfg := &backfiller.VaasBackfiller{
				LogLevel:          logLevel,
				P2pNetwork:        p2pNetwork,
				MongoURI:          mongoUri,
				MongoDatabase:     mongoDb,
				RequestsPerMinute: requestsPerMinute,
				StartTime:         startTime,
				EndTime:           endTime,
				PageSize:          pageSize,
				NumWorkers:        numWorkers,
				Overwrite:         overwrite,
				DisableDBUpsert:   disableDBUpsert,
				RpcProvidersPath:  rpcProvidersPath,
			}
			if emitterChainID != 0 {
				eci := sdk.ChainID(emitterChainID)
				cfg.EmitterChainID = &eci
			}
			if emitterAddress != "" {
				cfg.EmitterAddress = &emitterAddress
			}
			backfiller.RunByVaas(cfg)
		},
	}

	vaas.Flags().StringVar(&logLevel, "log-level", "INFO", "log level")
	vaas.Flags().StringVar(&p2pNetwork, "p2p-network", "", "P2P network to use")
	vaas.Flags().StringVar(&mongoUri, "mongo-uri", "", "Mongo connection")
	vaas.Flags().StringVar(&mongoDb, "mongo-database", "", "Mongo database")
	vaas.Flags().StringVar(&startTime, "start-time", "1970-01-01T00:00:00Z", "minimum VAA timestamp to process")
	vaas.Flags().StringVar(&endTime, "end-time", "", "maximum VAA timestamp to process (default now)")
	vaas.Flags().Int64Var(&pageSize, "page-size", 100, "number of documents retrieved at a time")
	vaas.Flags().Int64Var(&requestsPerMinute, "requests-per-minute", 12, "maximum number of requests per minute to process VAA documents")
	vaas.Flags().IntVar(&numWorkers, "num-workers", 1, "number of workers to process VAA documents concurrently")
	vaas.Flags().Uint16Var(&emitterChainID, "emitter-chain", 0, "emitter chain id")
	vaas.Flags().StringVar(&emitterAddress, "emitter-address", "", "emitter address")
	vaas.Flags().BoolVar(&overwrite, "overwrite", false, "overwrite existing data")
	vaas.Flags().BoolVar(&disableDBUpsert, "disable-db-upsert", false, "disable db upsert")
	vaas.Flags().StringVar(&rpcProvidersPath, "rpc-providers-path", "", "path to rpc providers file")

	vaas.MarkFlagRequired("mongo-uri")
	vaas.MarkFlagRequired("p2p-network")
	vaas.MarkFlagRequired("mongo-database")
	vaas.MarkFlagRequired("start-time")
	vaas.MarkFlagRequired("rpc-providers-path")

	parent.AddCommand(vaas)
}
