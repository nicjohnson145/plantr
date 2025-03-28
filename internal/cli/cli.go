package cli

import (
	"context"
	"fmt"
	"os"

	agentv1 "github.com/nicjohnson145/plantr/gen/plantr/agent/v1"
	"github.com/nicjohnson145/plantr/internal/agent"
	"github.com/nicjohnson145/plantr/internal/encryption"
	"github.com/rs/zerolog"
)

type CLIConfig struct {
	Logger     zerolog.Logger
	Agent      *agent.Agent
}

func NewCLI(conf CLIConfig) *CLI {
	return &CLI{
		log:   conf.Logger,
		agent: conf.Agent,
	}
}

type CLI struct {
	log        zerolog.Logger
	agent      *agent.Agent
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

func (c *CLI) Sync() error {
	_, err := c.agent.Sync(context.Background(), &agentv1.SyncRequest{})
	if err != nil {
		return fmt.Errorf("error syncing:\n%w", err)
	}
	return nil
}

func (c *CLI) ForceRefresh() error {
	return c.agent.ForceRefresh(context.Background())
}
