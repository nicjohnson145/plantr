package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/nicjohnson145/plantr/internal/config"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	_ "modernc.org/sqlite" // import sqlite driver
)

type Client interface {
	WriteChallenge(ctx context.Context, challenge *Challenge) error
	ReadChallenge(ctx context.Context, id string) (*Challenge, error)
}

func NewFromEnv(logger zerolog.Logger) (Client, func(), error) {
	cleanup := func() {}

	kind, err := config.ParseStorageKind(viper.GetString(config.StorageType))
	if err != nil {
		return nil, cleanup, err
	}

	var driver string
	var dsn string
	switch kind {
	case config.StorageKindSqlite:
		driver = "sqlite"
		dsn = viper.GetString(config.SqliteDBPath)
	default:
		return nil, cleanup, fmt.Errorf("unhandled type of '%v'", kind)
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
	case config.StorageKindSqlite:
		sqlite, err := NewSqlLite(SqlLiteConfig{
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
