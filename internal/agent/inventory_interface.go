package agent

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type InventoryClient interface {
	GetRow(ctx context.Context, hash string) (*InventoryRow, error)
	WriteRow(ctx context.Context, row InventoryRow) error
}

func NewInventoryClientFromEnv(logger zerolog.Logger) (InventoryClient, func(), error) {
	cleanup := func() {}

	kind, err := ParseStorageKind(viper.GetString(StorageType))
	if err != nil {
		return nil, cleanup, err
	}

	var driver string
	var dsn string
	switch kind {
	case StorageKindSqlite:
		driver = "sqlite"
		dsn = viper.GetString(SqliteDBPath)
		// ensure that any containing directories are created
		if err := os.MkdirAll(filepath.Dir(dsn), 0775); err != nil {
			return nil, cleanup, fmt.Errorf("error ensuring containing directories: %w", err)
		}
	case StorageKindNone:
		return NewNoopInventory(NoopInventoryConfig{Logger: logger}), func() {}, nil
	default:
		return nil, cleanup, fmt.Errorf("unhandled type of %v", kind)
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, cleanup, fmt.Errorf("error opening DB: %w", err)
	}

	cleanup = func() {
		db.Close()
	}

	logger.Info().Msg("executing migrations")
	if err := ExecuteMigrations(kind, db); err != nil {
		return nil, cleanup, fmt.Errorf("error executing migrations: %w", err)
	}

	switch kind {
	case StorageKindSqlite:
		sqlite, err := NewSqlLiteInventory(SqlLiteInventoryConfig{
			Logger: logger,
			DB:     db,
		})
		if err != nil {
			return nil, cleanup, fmt.Errorf("error initializing sqlite client: %w", err)
		}
		return sqlite, cleanup, nil
	default:
		return nil, cleanup, fmt.Errorf("unhandled type of '%v'", kind)
	}
}
