package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	PORT        string
	JWTSecret   string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load("/Users/manasagarwal/Documents/go/valet_clone/.env")
	if err != nil {
		return nil, err
	}

	var config *Config = &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		PORT:        os.Getenv("PORT"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
	}

	return config, nil
}
