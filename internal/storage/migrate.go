package storage

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/nicjohnson145/plantr/internal/config"
)

//go:embed sqlite-migrations/*.sql
var sqliteMigrations embed.FS

func ExecuteMigrations(dbType config.StorageKind, db *sql.DB) error {
	var driver database.Driver
	var kind string
	var fs embed.FS
	var path string

	switch dbType {
	case config.StorageKindSqlite:
		path = "sqlite-migrations"
		fs = sqliteMigrations
		kind = "sqlite"
		d, err := sqlite.WithInstance(db, &sqlite.Config{
			NoTxWrap: true,
		})
		if err != nil {
			return fmt.Errorf("error creating sqlite driver: %w", err)
		}
		driver = d
	default:
		return fmt.Errorf("unhandled db type of %v", dbType)
	}

	migrations, err := iofs.New(fs, path)
	if err != nil {
		return fmt.Errorf("error creating migrations source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", migrations, kind, driver)
	if err != nil {
		return fmt.Errorf("error creating migrations instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("error executing migrations: %w", err)
	}

	return nil
}
