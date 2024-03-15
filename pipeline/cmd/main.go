package main

import (
	"github.com/spf13/cobra"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/cmd/backfiller"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/cmd/service"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/config"
)

func main() {
	execute()
}

func execute() error {
	root := &cobra.Command{
		Use: "pipeline",
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
	var mongoUri, mongoDb, snsUrl, logLevel, awsRegion, startTime, endTime string
	var awsEndpoint, awsAccessKeyID, awsSecretAccessKey string
	var pageSize, requestsPerSecond int64
	var numWorkers int

	backfillerCommand := &cobra.Command{
		Use:   "backfiller",
		Short: "Run backfiller to send vaas to sns",
		Run: func(_ *cobra.Command, _ []string) {
			cfg := &config.Backfiller{
				LogLevel:           logLevel,
				MongoURI:           mongoUri,
				MongoDatabase:      mongoDb,
				AwsEndpoint:        awsEndpoint,
				AwsAccessKeyID:     awsAccessKeyID,
				AwsSecretAccessKey: awsSecretAccessKey,
				AwsRegion:          awsRegion,
				SNSUrl:             snsUrl,
				RequestsPerSecond:  requestsPerSecond,
				StartTime:          startTime,
				EndTime:            endTime,
				PageSize:           pageSize,
				NumWorkers:         numWorkers,
			}
			backfiller.Run(cfg)
		},
	}
	backfillerCommand.Flags().StringVar(&logLevel, "log-level", "INFO", "log level")
	backfillerCommand.Flags().StringVar(&mongoUri, "mongo-uri", "", "Mongo connection")
	backfillerCommand.Flags().StringVar(&mongoDb, "mongo-database", "", "Mongo database")
	backfillerCommand.Flags().StringVar(&snsUrl, "sns-url", "", "SNS Url topic to push vaas")
	backfillerCommand.Flags().StringVar(&awsRegion, "aws-region", "", "Aws region")
	backfillerCommand.Flags().StringVar(&awsEndpoint, "aws-endpoint", "", "Aws endpoint")
	backfillerCommand.Flags().StringVar(&awsAccessKeyID, "aws-access-key-id", "", "Aws access key id")
	backfillerCommand.Flags().StringVar(&awsSecretAccessKey, "aws-secret-access-key", "", "Aws secret access key")
	backfillerCommand.Flags().StringVar(&startTime, "start-time", "1970-01-01T00:00:00Z", "minimum VAA timestamp to process")
	backfillerCommand.Flags().StringVar(&endTime, "end-time", "", "maximum VAA timestamp to process (default now)")
	backfillerCommand.Flags().Int64Var(&pageSize, "page-size", 100, "number of documents retrieved at a time")
	backfillerCommand.Flags().Int64Var(&requestsPerSecond, "requests-per-second", 100, "maximum number of requests per second to publish to sns topic")
	backfillerCommand.Flags().IntVar(&numWorkers, "num-workers", 5, "number of workers to publish vaas")

	backfillerCommand.MarkFlagRequired("mongo-uri")
	backfillerCommand.MarkFlagRequired("mongo-database")
	backfillerCommand.MarkFlagRequired("aws-region")
	backfillerCommand.MarkFlagRequired("sns-url")
	backfillerCommand.MarkFlagRequired("start-time")

	root.AddCommand(backfillerCommand)
}
