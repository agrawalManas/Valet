package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDB(dbUrl string) (*pgxpool.Pool, error) {
	var ctx = context.Background()
	var err error
	var config *pgxpool.Config

	config, err = pgxpool.ParseConfig(dbUrl)
	if err != nil {
		return nil, err
	}

	var pool *pgxpool.Pool
	pool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		return nil, err
	}

	return pool, err
}
