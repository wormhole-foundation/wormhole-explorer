package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/xlabs/influx-backfiller/cmd/metrics"
	"github.com/xlabs/influx-backfiller/cmd/prices"
)

// Execute executes the root command.
func Execute() error {

	root := &cobra.Command{
		Use: "backfiller",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				_ = cmd.Help()
				os.Exit(0)
			}
		},
	}

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

	return root.Execute()
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
	//input flag
	vaaCountCmd.Flags().StringVar(&input, "input", "", "path to input vaa file")
	vaaCountCmd.MarkFlagRequired("input")
	//output flag
	vaaCountCmd.Flags().StringVar(&output, "output", "", "path to output file")
	vaaCountCmd.MarkFlagRequired("output")
	parent.AddCommand(vaaCountCmd)
}

func addVaaVolumeCommand(parent *cobra.Command) {
	var input, output string
	vaaVolumeCmd := &cobra.Command{
		Use:   "vaa-volume",
		Short: "Generate volume metrics from a VAA csv file",
		Run: func(_ *cobra.Command, _ []string) {
			metrics.RunVaaVolume(input, output)
		},
	}
	//input flag
	vaaVolumeCmd.Flags().StringVar(&input, "input", "", "path to input vaa file")
	vaaVolumeCmd.MarkFlagRequired("input")
	//output flag
	vaaVolumeCmd.Flags().StringVar(&output, "output", "", "path to output file")
	vaaVolumeCmd.MarkFlagRequired("output")
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
	//output flag
	vaaCountCmd.Flags().StringVar(&output, "output", "", "path to output file")
	vaaCountCmd.MarkFlagRequired("output")
	root.AddCommand(vaaCountCmd)
}
