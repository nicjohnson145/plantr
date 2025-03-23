package main

import (
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := root().Execute(); err != nil {
		os.Exit(1)
	}
}

func root() *cobra.Command {
	cmd := &cobra.Command{
		Use: "plantr",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// So we don't print usage messages on execution errors
			cmd.SilenceUsage = true
			// So we dont double report errors
			cmd.SilenceErrors = true
		},
	}

	cmd.AddCommand(
		generateKeyPair(),
		sync(),
	)

	return cmd
}
