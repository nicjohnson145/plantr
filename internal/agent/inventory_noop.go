package agent

import (
	"context"

	"github.com/rs/zerolog"
)

type NoopInventoryConfig struct {
	Logger zerolog.Logger
}

func NewNoopInventory(conf NoopInventoryConfig) *NoopInventory {
	return &NoopInventory{
		log: conf.Logger,
	}
}

type NoopInventory struct {
	log zerolog.Logger
}

func (n *NoopInventory) GetRow(ctx context.Context, hash string) (*InventoryRow, error) {
	return nil, nil
}

func (n *NoopInventory) WriteRow(ctx context.Context, row InventoryRow) error {
	return nil
}
