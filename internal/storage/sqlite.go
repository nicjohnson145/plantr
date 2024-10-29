package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/rs/zerolog"
	hsqlx "github.com/nicjohnson145/hlp/sqlx"
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
