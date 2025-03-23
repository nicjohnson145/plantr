package agent

import (
	"errors"
	"fmt"
	"os"

	"github.com/nicjohnson145/plantr/internal/logging"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

func NewAgentFromEnv(logger zerolog.Logger) (*Agent, func(), error) {
	cleanup := func() {}

	keyPath := viper.GetString(PrivateKeyPath)
	if keyPath == "" {
		return nil, cleanup, errors.New("private key path must be set")
	}
	privateKeyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, cleanup, fmt.Errorf("error reading private key: %w", err)
	}

	controllerAddress := viper.GetString(ControllerAddress)
	if controllerAddress == "" {
		return nil, cleanup, errors.New("controller address must be set")
	}

	nodeID := viper.GetString(NodeID)
	if nodeID == "" {
		return nil, cleanup, errors.New("node is must be set")
	}

	inventory, inventoryCleanup, err := NewInventoryClientFromEnv(logging.Component(logger, "inventory"))
	if err != nil {
		return nil, cleanup, fmt.Errorf("error creating inventory client: %w", err)
	}

	cleanup = inventoryCleanup

	return NewAgent(AgentConfig{
		Logger:            logging.Component(logger, "agent-worker"),
		NodeID:            nodeID,
		ControllerAddress: controllerAddress,
		PrivateKey:        string(privateKeyBytes),
		Inventory:         inventory,
	}), cleanup, nil
}
