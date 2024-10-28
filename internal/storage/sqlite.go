package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

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

func (s *SqlLite) RegisterHost(ctx context.Context, host *Host) (*Host, error) {
	return nil, nil
}
