package repository

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresConfig struct {
	ConnString      string
	MaxConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

func ConnectToPostgres(ctx context.Context, cfg *PostgresConfig) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.ConnString)
	if err != nil {
		return nil, err
	}

	// Set max open connections
	if cfg.MaxConns > 0 {
		db.SetMaxOpenConns(int(cfg.MaxConns))
	} else {
		db.SetMaxOpenConns(10)
	}

	// Set max connection lifetime
	if cfg.MaxConnLifetime > 0 {
		db.SetConnMaxLifetime(cfg.MaxConnLifetime)
	} else {
		db.SetConnMaxLifetime(time.Hour)
	}

	// Set max idle time
	if cfg.MaxConnIdleTime > 0 {
		db.SetConnMaxIdleTime(cfg.MaxConnIdleTime)
	} else {
		db.SetConnMaxIdleTime(5 * time.Minute)
	}

	// Ping to check connection
	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
