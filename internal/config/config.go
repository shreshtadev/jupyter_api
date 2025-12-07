package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr string // HTTP address
	DSN  string // MySQL/MariaDB DSN
}

func Load() *Config {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8382"
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN is required, e.g.: user:pass@tcp(localhost:3306)/dbname?parseTime=true&charset=utf8mb4&loc=Local")
	}

	return &Config{
		Addr: addr,
		DSN:  dsn,
	}
}

func LoadEnv() {
	_ = godotenv.Load(".env")
}
