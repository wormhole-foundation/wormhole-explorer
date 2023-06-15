package main

import (
	"os"

	"github.com/spf13/cobra"
)

func main() {
	execute()
}

func execute() error {
	root := &cobra.Command{
		Use: "backfiller",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				os.Exit(0)
			}
		},
	}

	addVaaBackfillerCommand(root)
	addTxHashCommand(root)
	addTxHashEncodingCommand(root)

	return root.Execute()
}

func addVaaBackfillerCommand(root *cobra.Command) {
	var mongoUri, mongoDb, filename string
	var workerCount int

	vaaBackfillerCommand := &cobra.Command{
		Use:   "vaa",
		Short: "Run vaa backfiller",
		Run: func(_ *cobra.Command, _ []string) {
			workerConfiguration := WorkerConfiguration{
				MongoURI:      mongoUri,
				MongoDatabase: mongoDb,
				Filename:      filename,
				WorkerCount:   workerCount,
			}
			RunBackfiller(workerConfiguration, workerVaa)
		},
	}
	vaaBackfillerCommand.Flags().StringVar(&mongoUri, "mongo-uri", "", "Mongo connection")
	vaaBackfillerCommand.Flags().StringVar(&mongoDb, "mongo-database", "", "Mongo database")
	vaaBackfillerCommand.Flags().StringVar(&filename, "filename", "", "vaa backfiller filename")
	vaaBackfillerCommand.Flags().IntVar(&workerCount, "worker-count", 100, "backfiller worker count")

	vaaBackfillerCommand.MarkFlagRequired("mongo-uri")
	vaaBackfillerCommand.MarkFlagRequired("mongo-database")
	vaaBackfillerCommand.MarkFlagRequired("filename")

	root.AddCommand(vaaBackfillerCommand)

}

func addTxHashCommand(root *cobra.Command) {
	var mongoUri, mongoDb, filename string
	var workerCount int

	txHashBackfillerCommand := &cobra.Command{
		Use:   "txhash",
		Short: "Run txhash backfiller",
		Run: func(_ *cobra.Command, _ []string) {
			workerConfiguration := WorkerConfiguration{
				MongoURI:      mongoUri,
				MongoDatabase: mongoDb,
				Filename:      filename,
				WorkerCount:   workerCount,
			}
			RunBackfiller(workerConfiguration, workerTxHash)
		},
	}
	txHashBackfillerCommand.Flags().StringVar(&mongoUri, "mongo-uri", "", "Mongo connection")
	txHashBackfillerCommand.Flags().StringVar(&mongoDb, "mongo-database", "", "Mongo database")
	txHashBackfillerCommand.Flags().StringVar(&filename, "filename", "", "vaa backfiller filename")
	txHashBackfillerCommand.Flags().IntVar(&workerCount, "worker-count", 100, "backfiller worker count")

	txHashBackfillerCommand.MarkFlagRequired("mongo-uri")
	txHashBackfillerCommand.MarkFlagRequired("mongo-database")
	txHashBackfillerCommand.MarkFlagRequired("filename")

	root.AddCommand(txHashBackfillerCommand)
}

func addTxHashEncodingCommand(root *cobra.Command) {
	var logLevel, mongoUri, mongoDb string
	var chainID uint16
	var pageSize int64
	txHashFixEncodingCommand := &cobra.Command{
		Use:   "txHashEncoding",
		Short: "Run txHash encoding backfiller",
		Run: func(_ *cobra.Command, _ []string) {
			cfg := TxHashEncondingConfig{
				LogLevel:      logLevel,
				MongoURI:      mongoUri,
				MongoDatabase: mongoDb,
				ChainID:       chainID,
				PageSize:      pageSize,
			}
			RunTxHashEncoding(cfg)
		},
	}

	txHashFixEncodingCommand.Flags().StringVar(&logLevel, "log-level", "info", "Log level")
	txHashFixEncodingCommand.Flags().StringVar(&mongoUri, "mongo-uri", "", "Mongo connection")
	txHashFixEncodingCommand.Flags().StringVar(&mongoDb, "mongo-database", "", "Mongo database")
	txHashFixEncodingCommand.Flags().Uint16Var(&chainID, "chain-id", 0, "Chain ID")
	txHashFixEncodingCommand.Flags().Int64Var(&pageSize, "page-size", 100, "Page size")

	txHashFixEncodingCommand.MarkFlagRequired("mongo-uri")
	txHashFixEncodingCommand.MarkFlagRequired("mongo-database")
	txHashFixEncodingCommand.MarkFlagRequired("chain-id")

	root.AddCommand(txHashFixEncodingCommand)
}
