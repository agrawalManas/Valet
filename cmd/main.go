package main

import (
	"agrawalManas/Valet/internal/config"
	"agrawalManas/Valet/internal/database"
	"agrawalManas/Valet/internal/router"
	"agrawalManas/Valet/internal/storage"
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	var cfg *config.Config
	var err error

	cfg, err = config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config", err)
	}

	//---------DB-Setup
	var pool *pgxpool.Pool
	pool, err = database.ConnectDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to load config", err)
	}
	defer pool.Close()

	//---------Storage-Setup
	var ctx = context.Background()
	r2Client, err := storage.NewR2Client(ctx)
	if err != nil {
		log.Fatalf("R2 initialization failed: %v", err)
	}

	if err := r2Client.CheckBucket(ctx); err != nil {
		log.Fatalf("R2 connection failed: %v", err)
	}

	//---------Router-Initialization
	if cfg.PORT == "" {
		log.Fatal("Not a valid port number " + cfg.PORT)
	}
	router.Initialize(cfg, pool, r2Client)
}
