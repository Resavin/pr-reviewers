package config

import (
	"os"
)

type Config struct {
	DBURL   string
	AppPort string
}

func Load() *Config {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080"
	}

	return &Config{
		DBURL:   dbURL,
		AppPort: appPort,
	}
}
