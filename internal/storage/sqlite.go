package storage

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/rs/zerolog"
)

type SqlLiteConfig struct {
	Logger zerolog.Logger
	DB     *sql.DB
}

func NewSqlLite(conf SqlLiteConfig) *SqlLite {
	return &SqlLite{
		log: conf.Logger,
		db:  sqlx.NewDb(conf.DB, "sqlite"),
	}
}

type SqlLite struct {
	log zerolog.Logger
	db  *sqlx.DB
}

func (s *SqlLite) RegisterHost(ctx context.Context, host *Host) (*Host, error) {
	return nil, nil
}
