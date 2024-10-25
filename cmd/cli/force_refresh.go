package main

import (
	"fmt"

	"github.com/nicjohnson145/plantr/internal/agent"
	"github.com/nicjohnson145/plantr/internal/cli"
	"github.com/nicjohnson145/plantr/internal/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func forceRefresh() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "force-refresh",
		Short: "Refresh controller info",
		Long: "Force the controller to refresh its copy of the seed repo",
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

			if err := c.ForceRefresh(); err != nil {
				logger.Err(err).Msg("error refreshing")
				return err
			}

			return nil
		},
	}

	return cmd
}
