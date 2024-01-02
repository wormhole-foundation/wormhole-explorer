package main

import (
	"github.com/spf13/cobra"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/cmd/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/cmd/prices"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/cmd/service"
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

func addVaaVolumeFromMongoCommand(parent *cobra.Command) {
	var mongoUri, mongoDb, output, prices, vaaPayloadParserURL, p2pNetwork string
	//vaa-volume from MongoDB
	vaaVolumeMongoCmd := &cobra.Command{
		Use:   "mongo",
		Short: "Generate volume metrics from MongoDB",
		Run: func(_ *cobra.Command, _ []string) {
			metrics.RunVaaVolumeFromMongo(mongoUri, mongoDb, output, prices, vaaPayloadParserURL, p2pNetwork)
		},
	}

	//mongo flags
	vaaVolumeMongoCmd.Flags().StringVar(&mongoUri, "mongo-uri", "", "Mongo connection")
	vaaVolumeMongoCmd.Flags().StringVar(&mongoDb, "mongo-database", "", "Mongo database")

	// output flag
	vaaVolumeMongoCmd.Flags().StringVar(&output, "output", "", "path to output file")
	vaaVolumeMongoCmd.MarkFlagRequired("output")
	// prices flag
	vaaVolumeMongoCmd.Flags().StringVar(&prices, "prices", "prices.csv", "path to prices file")

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
	addVaaVolumeFromMongoCommand(vaaVolumeCmd)
	parent.AddCommand(vaaVolumeCmd)
}

func addPricesCommand(root *cobra.Command) {
	var output, p2pNetwork, coingeckoUrl, coingeckoHeaderKey, coingeckoApiKey string
	vaaCountCmd := &cobra.Command{
		Use:   "history",
		Short: "Generate notional price history for symbol",
		Run: func(_ *cobra.Command, _ []string) {
			prices.RunPrices(output, p2pNetwork, coingeckoUrl, coingeckoHeaderKey, coingeckoApiKey)
		},
	}
	// output flag
	vaaCountCmd.Flags().StringVar(&output, "output", "", "path to output file")
	vaaCountCmd.MarkFlagRequired("output")

	//p2p-network flag
	vaaCountCmd.Flags().StringVar(&p2pNetwork, "p2p-network", "", "P2P network")
	vaaCountCmd.MarkFlagRequired("p2p-network")

	//coingecko flags
	vaaCountCmd.Flags().StringVar(&coingeckoUrl, "coingecko-url", "", "Coingecko URL")
	vaaCountCmd.MarkFlagRequired("coingecko-url")

	vaaCountCmd.Flags().StringVar(&coingeckoHeaderKey, "coingecko-header-key", "", "Coingecko header key")
	vaaCountCmd.Flags().StringVar(&coingeckoApiKey, "coingecko-api-key", "", "Coingecko api key")

	root.AddCommand(vaaCountCmd)
}
