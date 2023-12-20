package main

import (
	"github.com/spf13/cobra"
	"github.com/wormhole-foundation/wormhole-explorer/notional/cmd/service"
)

func main() {
	execute()
}

func execute() error {
	root := &cobra.Command{
		Use: "notional",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				service.Run()
			}
		},
	}

	return root.Execute()
}
