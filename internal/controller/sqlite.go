package controller

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	hsqlx "github.com/nicjohnson145/hlp/sqlx"
	"github.com/rs/zerolog"
)

type SqlLiteConfig struct {
	Logger zerolog.Logger
	DB     *sql.DB
}

func NewSqlLite(conf SqlLiteConfig) (*SqlLite, error) {
	sql := &SqlLite{
		log: conf.Logger,
		db:  sqlx.NewDb(conf.DB, "sqlite"),
	}
	if err := sql.init(); err != nil {
		return nil, err
	}

	return sql, nil
}

type SqlLite struct {
	log zerolog.Logger
	db  *sqlx.DB
}

func (s *SqlLite) init() error {
	if _, err := s.db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return fmt.Errorf("error enabling foreign key pragma: %w", err)
	}

	return nil
}

func (s *SqlLite) Purge(ctx context.Context) error {
	tables := []string{
		"challenge",
		"github_release_asset",
	}
	err := hsqlx.WithTransaction(s.db, func(txn *sqlx.Tx) error {
		for _, tbl := range tables {
			if _, err := txn.ExecContext(ctx, fmt.Sprintf("DELETE FROM %v", tbl)); err != nil {
				return fmt.Errorf("error deleting from %v: %w", tbl, err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error purging: %w", err)
	}
	return nil
}

func (s *SqlLite) WriteChallenge(ctx context.Context, challenge *Challenge) error {
	stmt := `
		INSERT INTO
			challenge
			(
				id,
				value
			)
		VALUES 
			(

				:id,
				:value
			)
	`

	if _, err := s.db.NamedExecContext(ctx, stmt, challenge); err != nil {
		return fmt.Errorf("error inserting: %w", err)
	}

	return nil
}

func (s *SqlLite) ReadChallenge(ctx context.Context, id string) (*Challenge, error) {
	stmt := `
		SELECT
			*
		FROM
			challenge
		WHERE
			id = :id
	`
	args := map[string]any{
		"id": id,
	}

	rows, err := hsqlx.RequireExactSelectNamedCtx[Challenge](ctx, 1, s.db, stmt, args)
	if err != nil {
		return nil, fmt.Errorf("error querying: %w", err)
	}

	return &rows[0], nil
}

func (s *SqlLite) ReadGithubRelease(ctx context.Context, release *DBGithubRelease) (string, error) {
	stmt := `
		SELECT
			download_url
		FROM
			github_release_cache
		WHERE
			repo = :repo AND
			tag = :tag AND
			os = :os AND
			arch = :arch
	`
	
	rows, err := hsqlx.RequireExactSelectNamedCtx[DBGithubRelease](ctx, 1, s.db, stmt, release)
	if err != nil {
		if errors.Is(err, hsqlx.ErrNotFoundError) {
			return "", nil
		}
		return "", fmt.Errorf("error selecting: %w", err)
	}
	return rows[0].DownloadURL, nil
}

func (s *SqlLite) WriteGithubReleaseAsset(ctx context.Context, release *DBGithubRelease) (error) {
	stmt := `
		INSERT OR REPLACE INTO
			github_release_cache
			(
				repo,
				tag,
				os,
				arch,
				download_url
			)
		VALUES
			(
				:repo,
				:tag,
				:os,
				:arch,
				:download_url
			)
	`
	if _, err := s.db.NamedExecContext(ctx, stmt, release); err != nil {
		return fmt.Errorf("error upserting cache: %w", err)
	}
	return nil
}

func (s *SqlLite) ReadGithubReleaseAsset(ctx context.Context, asset *DBGithubRelease) (string, error) {
	stmt := `
		SELECT
			*
		FROM
			github_release_asset
		WHERE
			hash = :hash AND
			os = :os AND
			arch = :arch
	`
	rows, err := hsqlx.RequireExactSelectNamedCtx[DBGithubRelease](ctx, 1, s.db, stmt, asset)
	if err != nil {
		if errors.Is(err, hsqlx.ErrNotFoundError) {
			return "", nil
		}
		return "", fmt.Errorf("error selecting: %w", err)
	}
	return rows[0].DownloadURL, nil
}
