package database

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/AdityaTote/wallet-service/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Pool *pgxpool.Pool
}

const DatabasePingTimeout = 3

func New(ctx context.Context, cfg config.Config) (*Database, error) {
	encodedPassword := url.QueryEscape(cfg.DbPassword)

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", cfg.DbHost, cfg.DbUser, encodedPassword, cfg.DbName, cfg.DbPort, "disable")

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pgx pool config: %w", err)
	}

	db, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, DatabasePingTimeout*time.Second)
	defer cancel()

	if err = db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &Database{
		Pool: db,
	}

	return database, nil
}

func (db *Database) Close() error {
	db.Close()
	return nil
}