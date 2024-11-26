package main

import (
	"github.com/spf13/cobra"
)

func generateKeyPair() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-keypair",
		Short: "Generate a node keypair",
		Long:  "Generate keys used for node authentication",
		RunE: func(cmd *cobra.Command, args []string) error {
			//config.InitConfig()

			//logger := logging.Init(&logging.LoggingConfig{
			//    Level:  logging.LogLevel(viper.GetString(config.LoggingLevel)),
			//    Format: logging.LogFormatHuman,
			//})

			//public, private, err := encryption.GenerateKeyPair(&encryption.KeyOpts{})
			//if err != nil {
			//    logger.Err(err).Msg("error generating keypair")
			//    return err
			//}

			//if err := os.WriteFile("key", []byte(private), 0664); err != nil { //nolint: gosec
			//    logger.Err(err).Msg("error writing private key file")
			//    return err
			//}

			//if err := os.WriteFile("key.pub", []byte(public), 0664); err != nil { //nolint: gosec
			//    logger.Err(err).Msg("error writing public key file")
			//    return err
			//}

			return nil
		},
	}

	return cmd
}
