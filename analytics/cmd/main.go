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

func addVaaVolumeCommand(parent *cobra.Command) {
	var input, output, prices string
	vaaVolumeCmd := &cobra.Command{
		Use:   "vaa-volume",
		Short: "Generate volume metrics from a VAA csv file",
		Run: func(_ *cobra.Command, _ []string) {
			metrics.RunVaaVolume(input, output, prices)
		},
	}
	// input flag
	vaaVolumeCmd.Flags().StringVar(&input, "input", "", "path to input vaa file")
	vaaVolumeCmd.MarkFlagRequired("input")
	// output flag
	vaaVolumeCmd.Flags().StringVar(&output, "output", "", "path to output file")
	vaaVolumeCmd.MarkFlagRequired("output")
	// prices flag
	vaaVolumeCmd.Flags().StringVar(&prices, "prices", "prices.csv", "path to prices file")

	parent.AddCommand(vaaVolumeCmd)
}

func addPricesCommand(root *cobra.Command) {
	var output string
	vaaCountCmd := &cobra.Command{
		Use:   "history",
		Short: "Generate notional price history for symbol",
		Run: func(_ *cobra.Command, _ []string) {
			prices.RunPrices(output)
		},
	}
	// output flag
	vaaCountCmd.Flags().StringVar(&output, "output", "", "path to output file")
	vaaCountCmd.MarkFlagRequired("output")
	root.AddCommand(vaaCountCmd)
}
