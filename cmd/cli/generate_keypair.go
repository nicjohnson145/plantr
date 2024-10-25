package main

import (
	"fmt"

	"github.com/nicjohnson145/plantr/internal/cli"
	"github.com/nicjohnson145/plantr/internal/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func generateKeyPair() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-keypair",
		Short: "Generate a node keypair",
		Long:  "Generate keys used for node authentication",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cli.InitConfig(); err != nil {
				fmt.Printf("error initializing config: %v\n", err)
				return err
			}

			logger := logging.Init(&logging.LoggingConfig{
				Level:  logging.LogLevel(viper.GetString(cli.LoggingLevel)),
				Format: logging.LogFormat(viper.GetString(cli.LoggingFormat)),
			})

			c := cli.NewCLI(cli.CLIConfig{
				Logger: logger,
			})

			if err := c.GenerateKeyPair(); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
