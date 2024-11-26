package controller

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	_ "modernc.org/sqlite" // import sqlite driver
)

type StorageClient interface {
	WriteChallenge(ctx context.Context, challenge *Challenge) error
	ReadChallenge(ctx context.Context, id string) (*Challenge, error)
}

func NewStorageClientFromEnv(logger zerolog.Logger) (StorageClient, func(), error) {
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
	case StorageKindSqlite:
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
