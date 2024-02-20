package main

import (
	"github.com/spf13/cobra"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/cmd/backfiller"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/cmd/service"
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

	addBackfillerByTimeRange(backfiller)
	addBackfillerForIncompletes(backfiller)
	addBackfillerByVaas(backfiller)
	parent.AddCommand(backfiller)

}

func addBackfillerByTimeRange(parent *cobra.Command) {
	var before, after string
	timeRange := &cobra.Command{
		Use:   "time-range",
		Short: "Run backfiller for a time range",
		Run: func(_ *cobra.Command, _ []string) {
			backfiller.RunByTimeRange(after, before)
		},
	}
	// before flag
	timeRange.Flags().StringVar(&before, "before", "", "path to input vaa file")
	timeRange.MarkFlagRequired("before")
	// after flag
	timeRange.Flags().StringVar(&after, "after", "", "path to output file")
	timeRange.MarkFlagRequired("after")
	parent.AddCommand(timeRange)
}

func addBackfillerForIncompletes(parent *cobra.Command) {
	incompletes := &cobra.Command{
		Use:   "incompletes",
		Short: "Run backfiller for source tx incompletes",
		Run: func(_ *cobra.Command, _ []string) {
			backfiller.RunForIncompletes()
		},
	}
	parent.AddCommand(incompletes)
}

func addBackfillerByVaas(parent *cobra.Command) {
	var emitterAddress, sequence string
	var emitterChainID uint16
	vaas := &cobra.Command{
		Use:   "vaas",
		Short: "Run backfiller for vaas",
		Run: func(_ *cobra.Command, _ []string) {
			backfiller.RunByVaas(emitterChainID, emitterAddress, sequence)
		},
	}
	// emitter-chain flag
	vaas.Flags().Uint16Var(&emitterChainID, "emitter-chain", 0, "path to input vaa file")
	vaas.MarkFlagRequired("emitter-chain")

	// emitter-address flag
	vaas.Flags().StringVar(&emitterAddress, "emitter-address", "", "path to output file")

	// sequence flag
	vaas.Flags().StringVar(&sequence, "sequence", "", "path to output file")

	parent.AddCommand(vaas)
}
