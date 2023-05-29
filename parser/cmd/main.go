package main

import (
	"github.com/spf13/cobra"
	"github.com/wormhole-foundation/wormhole-explorer/parser/cmd/backfiller"
	"github.com/wormhole-foundation/wormhole-explorer/parser/cmd/service"
	"github.com/wormhole-foundation/wormhole-explorer/parser/config"
)

func main() {
	execute()
}

func execute() error {
	root := &cobra.Command{
		Use: "parser",
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
		Short: "Run parser as service",
		Run: func(_ *cobra.Command, _ []string) {
			service.Run()
		},
	}
	root.AddCommand(serviceCommand)
}

func addBackfiller(root *cobra.Command) {
	var mongoUri, mongoDb, vaaPayloadParserURL, logLevel, startTime, endTime string
	var vaaPayloadParserTimeout, pageSize int64

	backfillerCommand := &cobra.Command{
		Use:   "backfiller",
		Short: "Run backfiller to backfill data",
		Run: func(_ *cobra.Command, _ []string) {
			cfg := &config.BackfillerConfiguration{
				LogLevel:                logLevel,
				MongoURI:                mongoUri,
				MongoDatabase:           mongoDb,
				VaaPayloadParserURL:     vaaPayloadParserURL,
				VaaPayloadParserTimeout: vaaPayloadParserTimeout,
				StartTime:               startTime,
				EndTime:                 endTime,
				PageSize:                pageSize,
			}
			backfiller.Run(cfg)
		},
	}
	backfillerCommand.Flags().StringVar(&logLevel, "log-level", "INFO", "log level")
	backfillerCommand.Flags().StringVar(&mongoUri, "mongo-uri", "", "Mongo connection")
	backfillerCommand.Flags().StringVar(&mongoDb, "mongo-database", "", "Mongo database")
	backfillerCommand.Flags().StringVar(&vaaPayloadParserURL, "vaa-payload-parser-url", "", "VAA payload parser service URL")
	backfillerCommand.Flags().Int64Var(&vaaPayloadParserTimeout, "vaa-payload-parser-timeout", 10, "maximum waiting time in call to VAA payload service in seconds")
	backfillerCommand.Flags().StringVar(&startTime, "start-time", "1970-01-01T00:00:00Z", "minimum VAA timestamp to process")
	backfillerCommand.Flags().StringVar(&endTime, "end-time", "", "maximum VAA timestamp to process (default now)")
	backfillerCommand.Flags().Int64Var(&pageSize, "page-size", 100, "number of documents retrieved at a time")

	backfillerCommand.MarkFlagRequired("mongo-uri")
	backfillerCommand.MarkFlagRequired("mongo-database")
	backfillerCommand.MarkFlagRequired("p2p-network")
	backfillerCommand.MarkFlagRequired("vaa-payload-parser-url")

	root.AddCommand(backfillerCommand)
}
