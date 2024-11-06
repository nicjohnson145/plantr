package main

import (
	"os"

	"github.com/nicjohnson145/plantr/internal/config"
	"github.com/nicjohnson145/plantr/internal/encryption"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func generateKeyPair() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-keypair",
		Short: "Generate a node keypair",
		Long:  "Generate keys used for node authentication",
		RunE: func(cmd *cobra.Command, args []string) error {
			config.InitConfig()

			logger := config.Init(&config.LoggingConfig{
				Level:  config.LogLevel(viper.GetString(config.LoggingLevel)),
				Format: config.LogFormatHuman,
			})

			public, private, err := encryption.GenerateKeyPair(&encryption.KeyOpts{})
			if err != nil {
				logger.Err(err).Msg("error generating keypair")
				return err
			}

			if err := os.WriteFile("key", []byte(private), 0664); err != nil {
				logger.Err(err).Msg("error writing private key file")
				return err
			}

			if err := os.WriteFile("key.pub", []byte(public), 0664); err != nil {
				logger.Err(err).Msg("error writing public key file")
				return err
			}

			return nil
		},
	}

	return cmd
}
