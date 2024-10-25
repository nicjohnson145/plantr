package main

import (
	"fmt"

	"github.com/nicjohnson145/plantr/internal/agent"
	"github.com/nicjohnson145/plantr/internal/cli"
	"github.com/nicjohnson145/plantr/internal/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func sync() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cli.InitConfig(); err != nil {
				fmt.Printf("error initializing config: %v\n", err)
				return err
			}
			logger := logging.Init(&logging.LoggingConfig{
				Level:  logging.LogLevel(viper.GetString(cli.LoggingLevel)),
				Format: logging.LogFormat(viper.GetString(cli.LoggingFormat)),
			})

			worker, workerCleanup, err := agent.NewAgentFromEnv(logger)
			if err != nil {
				logger.Err(err).Msg("error creating agent")
				return err
			}
			defer workerCleanup()

			c := cli.NewCLI(cli.CLIConfig{
				Logger: logger,
				Agent:  worker,
			})

			if err := c.Sync(); err != nil {
				logger.Error().Msg("error executing sync")
				// Print this, cause it will likely have multi-line errors in it
				fmt.Println(err)
				return err
			}

			return nil
		},
	}

	return cmd
}
