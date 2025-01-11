package main

import (
	"github.com/nicjohnson145/plantr/internal/cli"
	"github.com/nicjohnson145/plantr/internal/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize node",
		Long:  "Setup node-level configuration for cli/agent usage",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli.InitConfig()

			logger := logging.Init(&logging.LoggingConfig{
				Level:  logging.LogLevel(viper.GetString(cli.LoggingLevel)),
				Format: logging.LogFormat(viper.GetString(cli.LoggingFormat)),
			})

			c := cli.NewCLI(cli.CLIConfig{
				Logger: logger,
			})

			opts := cli.InitOpts{
				ControllerAddress: viper.GetString(cli.InitControllerAddress),
				ID:                viper.GetString(cli.InitNodeID),
				PublicKeyPath:     viper.GetString(cli.InitPublicKeyPath),
				UserHome:          viper.GetString(cli.InitUserHome),
				PackageManager:    viper.GetString(cli.InitPackageManager),
			}

			if err := c.Init(opts); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
