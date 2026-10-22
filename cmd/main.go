package main

import (
	"agrawalManas/Valet/internal/config"
	"agrawalManas/Valet/internal/database"
	"agrawalManas/Valet/internal/router"
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

	//---------Router-Initialization
	if cfg.PORT == "" {
		log.Fatal("Not a valid port number " + cfg.PORT)
	}
	router.Initialize(cfg, pool)
}
