package cli

import (
	"os"

	"github.com/nicjohnson145/plantr/internal/encryption"
	"github.com/rs/zerolog"
)

type CLIConfig struct {
	Logger zerolog.Logger
}

func NewCLI(conf CLIConfig) *CLI {
	return &CLI{
		log: conf.Logger,
	}
}

type CLI struct {
	log zerolog.Logger
}

func (c *CLI) GenerateKeyPair() error {
	public, private, err := encryption.GenerateKeyPair(&encryption.KeyOpts{})
	if err != nil {
		c.log.Err(err).Msg("error generating keypair")
		return err
	}

	if err := os.WriteFile("key", []byte(private), 0664); err != nil { //nolint: gosec
		c.log.Err(err).Msg("error writing private key file")
		return err
	}

	if err := os.WriteFile("key.pub", []byte(public), 0664); err != nil { //nolint: gosec
		c.log.Err(err).Msg("error writing public key file")
		return err
	}

	return nil
}
