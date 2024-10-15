package main

import (
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/cmd/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/cmd/prices"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/cmd/service"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

func main() {
	execute()
}

func execute() error {
	root := &cobra.Command{
		Use: "analytics",
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
		Short: "Run analytics as service",
		Run: func(_ *cobra.Command, _ []string) {
			service.Run()
		},
	}
	root.AddCommand(serviceCommand)
}

func addBackfiller(root *cobra.Command) {
	metrics := &cobra.Command{
		Use: "metrics",
	}
	addVaaCountCommand(metrics)
	addVaaVolumeCommand(metrics)
	root.AddCommand(metrics)

	prices := &cobra.Command{
		Use: "prices",
	}
	addPricesCommand(prices)
	root.AddCommand(prices)
}

func addVaaCountCommand(parent *cobra.Command) {
	var input, output string
	vaaCountCmd := &cobra.Command{
		Use:   "vaa-count",
		Short: "Generate vaa-count metrics from a vaa csv file",
		Run: func(_ *cobra.Command, _ []string) {
			metrics.RunVaaCount(input, output)
		},
	}
	// input flag
	vaaCountCmd.Flags().StringVar(&input, "input", "", "path to input vaa file")
	vaaCountCmd.MarkFlagRequired("input")
	// output flag
	vaaCountCmd.Flags().StringVar(&output, "output", "", "path to output file")
	vaaCountCmd.MarkFlagRequired("output")
	parent.AddCommand(vaaCountCmd)
}

func addVaaVolumeFromFileCommand(parent *cobra.Command) {
	var input, output, prices, vaaPayloadParserURL, p2pNetwork string

	//vaa-volume from csv file
	vaaVolumeFileCmd := &cobra.Command{
		Use:   "file",
		Short: "Generate volume metrics from a VAA csv file",
		Run: func(_ *cobra.Command, _ []string) {
			metrics.RunVaaVolumeFromFile(input, output, prices, vaaPayloadParserURL, p2pNetwork)
		},
	}

	// input flag
	vaaVolumeFileCmd.Flags().StringVar(&input, "input", "", "path to input vaa file")
	vaaVolumeFileCmd.MarkFlagRequired("input")
	// output flag
	vaaVolumeFileCmd.Flags().StringVar(&output, "output", "", "path to output file")
	vaaVolumeFileCmd.MarkFlagRequired("output")
	// prices flag
	vaaVolumeFileCmd.Flags().StringVar(&prices, "prices", "prices.csv", "path to prices file")

	//vaa-payload-parser-url flag
	vaaVolumeFileCmd.Flags().StringVar(&vaaPayloadParserURL, "vaa-payload-parser-url", "", "VAA payload parser URL")
	vaaVolumeFileCmd.MarkFlagRequired("vaa-payload-parser-url")

	//p2p-network flag
	vaaVolumeFileCmd.Flags().StringVar(&p2pNetwork, "p2p-network", "", "P2P network")
	vaaVolumeFileCmd.MarkFlagRequired("p2p-network")

	parent.AddCommand(vaaVolumeFileCmd)
}

func addVaaVolumeFromDBCommand(parent *cobra.Command) {
	var dbLayer, mongoUri, mongoDb, dbURL, output, prices, vaaPayloadParserURL, p2pNetwork string
	//vaa-volume from MongoDB
	vaaVolumeDbCmd := &cobra.Command{
		Use:   "mongo",
		Short: "Generate volume metrics from db",
		Run: func(_ *cobra.Command, _ []string) {
			cfg := metrics.VaasVolume{
				DbLayer:             dbLayer,
				MongoUri:            mongoUri,
				MongoDb:             mongoDb,
				DbURL:               dbURL,
				OutputFile:          output,
				PricesFile:          prices,
				VaaPayloadParserUrl: vaaPayloadParserURL,
				P2PNetwork:          p2pNetwork,
			}
			metrics.RunVaaVolumeFromDb(cfg)
		},
	}

	//mongo flags
	vaaVolumeDbCmd.Flags().StringVar(&dbLayer, "db-layer", "mongo", "DB connection")
	vaaVolumeDbCmd.Flags().StringVar(&mongoUri, "mongo-uri", "", "Mongo connection")
	vaaVolumeDbCmd.Flags().StringVar(&mongoDb, "mongo-database", "", "Mongo database")
	vaaVolumeDbCmd.Flags().StringVar(&dbURL, "db-url", "", "DB connection")

	// output flag
	vaaVolumeDbCmd.Flags().StringVar(&output, "output", "", "path to output file")
	vaaVolumeDbCmd.MarkFlagRequired("output")
	// prices flag
	vaaVolumeDbCmd.Flags().StringVar(&prices, "prices", "prices.csv", "path to prices file")

	//vaa-payload-parser-url flag
	vaaVolumeDbCmd.Flags().StringVar(&vaaPayloadParserURL, "vaa-payload-parser-url", "", "VAA payload parser URL")
	vaaVolumeDbCmd.MarkFlagRequired("vaa-payload-parser-url")

	//p2p-network flag
	vaaVolumeDbCmd.Flags().StringVar(&p2pNetwork, "p2p-network", "", "P2P network")
	vaaVolumeDbCmd.MarkFlagRequired("p2p-network")

	parent.AddCommand(vaaVolumeDbCmd)

}

func addVaaVolumeV3FromVaasCollectionDump(parent *cobra.Command) {
	var vaasBsonFile, output, prices, vaaPayloadParserURL, p2pNetwork string
	//vaa-volume from MongoDB
	vaaVolumeMongoCmd := &cobra.Command{
		Use:   "vaa-volume-v3",
		Short: "Generate volume metrics from MongoDB vaas collection dump",
		Run: func(_ *cobra.Command, _ []string) {
			metrics.RunBackFillerVaaVolumeV3(vaasBsonFile, output, prices, vaaPayloadParserURL, p2pNetwork)
		},
	}

	// vaas bson file
	vaaVolumeMongoCmd.Flags().StringVar(&vaasBsonFile, "vaas-bson-file", "vaas.bson", "path to vaas bson file")
	vaaVolumeMongoCmd.MarkFlagRequired("vaas-bson-file")

	// output flag
	vaaVolumeMongoCmd.Flags().StringVar(&output, "output", "", "path to output file")
	vaaVolumeMongoCmd.MarkFlagRequired("output")

	// prices flag
	vaaVolumeMongoCmd.Flags().StringVar(&prices, "prices", "prices.csv", "path to prices file")
	vaaVolumeMongoCmd.MarkFlagRequired("prices")

	//vaa-payload-parser-url flag
	vaaVolumeMongoCmd.Flags().StringVar(&vaaPayloadParserURL, "vaa-payload-parser-url", "", "VAA payload parser URL")
	vaaVolumeMongoCmd.MarkFlagRequired("vaa-payload-parser-url")

	//p2p-network flag
	vaaVolumeMongoCmd.Flags().StringVar(&p2pNetwork, "p2p-network", "", "P2P network")
	vaaVolumeMongoCmd.MarkFlagRequired("p2p-network")

	parent.AddCommand(vaaVolumeMongoCmd)
}

func addVaaVolumeCommand(parent *cobra.Command) {

	vaaVolumeCmd := &cobra.Command{
		Use:   "vaa-volume",
		Short: "Generate volume metric",
	}

	addVaaVolumeFromFileCommand(vaaVolumeCmd)
	addVaaVolumeFromDBCommand(vaaVolumeCmd)
	addVaaVolumeV3FromVaasCollectionDump(vaaVolumeCmd)
	parent.AddCommand(vaaVolumeCmd)
}

func addPricesCommand(parent *cobra.Command) {
	addHistoryPrices(parent)
	addVaasPrices(parent)
}

func addHistoryPrices(parent *cobra.Command) {
	var output, p2pNetwork, coingeckoUrl, coingeckoHeaderKey, coingeckoApiKey string
	historyPricesCmd := &cobra.Command{
		Use:   "history",
		Short: "Generate notional price history for symbol",
		Run: func(_ *cobra.Command, _ []string) {
			prices.RunHistoryPrices(output, p2pNetwork, coingeckoUrl, coingeckoHeaderKey, coingeckoApiKey)
		},
	}
	// output flag
	historyPricesCmd.Flags().StringVar(&output, "output", "", "path to output file")
	historyPricesCmd.MarkFlagRequired("output")

	//p2p-network flag
	historyPricesCmd.Flags().StringVar(&p2pNetwork, "p2p-network", "", "P2P network")
	historyPricesCmd.MarkFlagRequired("p2p-network")

	//coingecko flags
	historyPricesCmd.Flags().StringVar(&coingeckoUrl, "coingecko-url", "", "Coingecko URL")
	historyPricesCmd.MarkFlagRequired("coingecko-url")

	historyPricesCmd.Flags().StringVar(&coingeckoHeaderKey, "coingecko-header-key", "", "Coingecko header key")
	historyPricesCmd.Flags().StringVar(&coingeckoApiKey, "coingecko-api-key", "", "Coingecko api key")

	parent.AddCommand(historyPricesCmd)
}

func addVaasPrices(parent *cobra.Command) {
	var cfg prices.VaasPrices
	var start, end, emitterAddress, sequence string
	var emitterChainID uint16
	vaasPricesCmd := &cobra.Command{
		Use:   "vaas",
		Short: "Add price to VAA",
		Run: func(_ *cobra.Command, _ []string) {
			if emitterChainID != 0 {
				eci := sdk.ChainID(emitterChainID)
				cfg.EmitterChainID = &eci
			}
			if emitterAddress != "" {
				cfg.EmitterAddress = &emitterAddress
			}
			if sequence != "" {
				cfg.Sequence = &sequence
			}
			if start != "" {
				st, err := time.Parse(time.RFC3339, start)
				if err != nil {
					log.Fatal("Failed to parse start: ", err)
				}
				cfg.StartTime = &st
			}
			if end != "" {
				et, err := time.Parse(time.RFC3339, end)
				if err != nil {
					log.Fatal("Failed to parse end: ", err)
				}
				cfg.StartTime = &et
			}
			prices.RunVaasPrices(cfg)
		},
	}

	//db flags
	vaasPricesCmd.Flags().StringVar(&cfg.DbLayer, "db-layer", "mongo", "DB connection")
	vaasPricesCmd.Flags().StringVar(&cfg.MongoUri, "mongo-uri", "", "Mongo connection")
	vaasPricesCmd.Flags().StringVar(&cfg.MongoDb, "mongo-database", "", "Mongo database")
	vaasPricesCmd.Flags().Int64Var(&cfg.PageSize, "page-size", 1000, "number of documents retrieved at a time")
	vaasPricesCmd.Flags().StringVar(&cfg.DbURL, "db-url", "", "DB connection")

	//p2p-network flag
	vaasPricesCmd.Flags().StringVar(&cfg.P2PNetwork, "p2p-network", "", "P2P network")
	vaasPricesCmd.MarkFlagRequired("p2p-network")

	//notional url flags
	vaasPricesCmd.Flags().StringVar(&cfg.NotionalUrl, "notional-url", "", "Notional URL")
	vaasPricesCmd.MarkFlagRequired("notional-url")

	//vaa-payload-parser-url flag
	vaasPricesCmd.Flags().StringVar(&cfg.VaaPayloadParserUrl, "vaa-payload-parser-url", "", "VAA payload parser URL")
	vaasPricesCmd.MarkFlagRequired("vaa-payload-parser-url")

	// emitter-chain flag
	vaasPricesCmd.Flags().Uint16Var(&emitterChainID, "emitter-chain", 0, "emitter chain id")

	// emitter-address flag
	vaasPricesCmd.Flags().StringVar(&emitterAddress, "emitter-address", "", "emitter address")

	// sequence flag
	vaasPricesCmd.Flags().StringVar(&sequence, "sequence", "", "sequence")

	// start flag
	vaasPricesCmd.Flags().StringVar(&start, "start", "", "start timestamp in RFC3339 format")

	// end flag
	vaasPricesCmd.Flags().StringVar(&end, "end", "", "end timestamp in RFC3339 format")

	parent.AddCommand(vaasPricesCmd)
}
