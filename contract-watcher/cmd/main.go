package main

import (
	"github.com/spf13/cobra"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/cmd/backfiller"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/cmd/service"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/config"
)

func main() {
	execute()
}

func execute() error {
	root := &cobra.Command{
		Use: "contract-watcher",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				service.Run()
			}
		},
	}

	addServiceCommand(root)
	addBackfillerCommand(root)

	return root.Execute()
}

func addServiceCommand(root *cobra.Command) {
	serviceCommand := &cobra.Command{
		Use:   "service",
		Short: "Run contract-watcher as service",
		Run: func(_ *cobra.Command, _ []string) {
			service.Run()
		},
	}
	root.AddCommand(serviceCommand)
}

func addBackfillerCommand(parent *cobra.Command) {
	var network, mongoUri, mongoDb, chainName, chainURL, logLevel string
	var fromBlock, toBlock, pageSize uint64
	var rateLimit int
	var persistBlock bool
	backfillerCommand := &cobra.Command{
		Use:   "backfiller",
		Short: "Run backfiller to backfill data",
		Run: func(c *cobra.Command, _ []string) {
			cfg := &config.BackfillerConfiguration{
				LogLevel:           logLevel,
				Network:            network,
				MongoURI:           mongoUri,
				MongoDatabase:      mongoDb,
				ChainName:          chainName,
				ChainUrl:           chainURL,
				FromBlock:          fromBlock,
				ToBlock:            toBlock,
				RateLimitPerSecond: rateLimit,
				PageSize:           pageSize,
				PersistBlock:       persistBlock,
			}

			backfiller.Run(cfg)
		},
	}
	backfillerCommand.Flags().StringVar(&logLevel, "log-level", "INFO", "log level")
	backfillerCommand.Flags().StringVar(&network, "network", "", "network (mainnet or testnet)")
	backfillerCommand.Flags().StringVar(&mongoUri, "mongo-uri", "", "Mongo connection")
	backfillerCommand.Flags().StringVar(&mongoDb, "mongo-database", "", "Mongo database")
	backfillerCommand.Flags().StringVar(&chainName, "chain-name", "", "chain name")
	backfillerCommand.Flags().StringVar(&chainURL, "chain-url", "", "chain URL")
	backfillerCommand.Flags().Uint64Var(&fromBlock, "from", 0, "first block to be processed")
	backfillerCommand.Flags().Uint64Var(&toBlock, "to", 0, "last block to be processed (included)")
	backfillerCommand.Flags().IntVar(&rateLimit, "rate-limit", 3, "rate limit per second")
	backfillerCommand.Flags().Uint64Var(&pageSize, "page-size", 100, "maximum number to process at one time")
	backfillerCommand.Flags().BoolVar(&persistBlock, "persist-blocks", false, "persist processed blocks in storage")

	backfillerCommand.MarkFlagRequired("network")
	backfillerCommand.MarkFlagRequired("mongo-uri")
	backfillerCommand.MarkFlagRequired("mongo-database")
	backfillerCommand.MarkFlagRequired("chain-name")
	backfillerCommand.MarkFlagRequired("chain-url")
	backfillerCommand.MarkFlagRequired("from")
	backfillerCommand.MarkFlagRequired("to")

	parent.AddCommand(backfillerCommand)
}
