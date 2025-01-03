package agent

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/nicjohnson145/hlp"
	hsqlx "github.com/nicjohnson145/hlp/sqlx"
	"github.com/rs/zerolog"
)

type SqlLiteInventoryConfig struct {
	Logger zerolog.Logger
	DB     *sql.DB
}

func NewSqlLiteInventory(conf SqlLiteInventoryConfig) (*SqlLiteInventory, error) {
	cli :=  &SqlLiteInventory{
		log: conf.Logger,
		db:  sqlx.NewDb(conf.DB, "sqlite"),
	}

	if err := cli.init(); err != nil {
		return nil, err
	}

	return cli, nil
}

type SqlLiteInventory struct {
	log zerolog.Logger
	db  *sqlx.DB
}

func (s *SqlLiteInventory) init() error {
	if _, err := s.db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return fmt.Errorf("error enabling foreign key pragma: %w", err)
	}

	return nil
}

func (s *SqlLiteInventory) GetRow(ctx context.Context, hash string) (*InventoryRow, error) {
	stmt := `
		SELECT
			*
		FROM
			agent_inventory
		WHERE
			hash = :hash
	`
	args := map[string]any{
		"hash": hash,
	}

	rows, err := hsqlx.RequireExactSelectNamedCtx[DBInventoryRow](ctx, 1, s.db, stmt, args)
	if err != nil {
		if errors.Is(err, hsqlx.ErrNotFoundError) {
			return nil, nil
		}
		return nil, fmt.Errorf("error selecting: %w", err)
	}

	return hlp.Ptr(rows[0].ToInventoryRow()), nil
}

func (s *SqlLiteInventory) WriteRow(ctx context.Context, row InventoryRow) error {
	return hsqlx.WithTransaction(s.db, func(txn *sqlx.Tx) error {
		// Start by purging old rows, since if we're writing then we've overwritten them
		if row.Package != nil {
			if err := s.purgeByPackage(ctx, txn, *row.Package); err != nil {
				return fmt.Errorf("error purging old package rows: %w", err)
			}
		}
		if row.Path != nil {
			if err := s.purgeByPath(ctx, txn, *row.Path); err != nil {
				return fmt.Errorf("error purging old path rows: %w", err)
			}
		}

		// Then insert our new one
		stmt := `
			INSERT INTO
				agent_inventory
				(
					hash,
					path,
					package
				)
			VALUES
				(
					:hash,
					:path,
					:package
				)
		`

		if _, err := txn.NamedExecContext(ctx, stmt, row.ToDBRow()); err != nil {
			return fmt.Errorf("error inserting: %w", err)
		}

		return nil
	})
}

func (s *SqlLiteInventory) purgeByPath(ctx context.Context, txn *sqlx.Tx, path string) error {
	return s.purgeByColumn(ctx, txn, "path", path)
}

func (s *SqlLiteInventory) purgeByPackage(ctx context.Context, txn *sqlx.Tx, pkg string) error {
	return s.purgeByColumn(ctx, txn, "package", pkg)
}

func (s *SqlLiteInventory) purgeByColumn(ctx context.Context, txn *sqlx.Tx, column string, value string) error {
	stmt := fmt.Sprintf(`
		DELETE FROM
			agent_inventory
		WHERE
			%v = :val
	`, column)

	args := map[string]any{
		"val": value,
	}

	if _, err := txn.NamedExecContext(ctx, stmt, args); err != nil {
		return fmt.Errorf("error deleting: %w", err)
	}

	return nil
}
