package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr           string // HTTP address
	DSN            string // MySQL/MariaDB DSN
	PrivateKeyPath string
	PublicKeyPath  string
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

	privateKey := os.Getenv("PRIVATE_KEY_PATH")
	if privateKey == "" {
		log.Fatal("PRIVATE_KEY_PATH is required.")
	}

	publicKey := os.Getenv("PUBLIC_KEY_PATH")
	if publicKey == "" {
		log.Fatal("PUBLIC_KEY_PATH is required.")
	}

	return &Config{
		Addr:           addr,
		DSN:            dsn,
		PrivateKeyPath: privateKey,
		PublicKeyPath:  publicKey,
	}
}

func LoadEnv() {
	_ = godotenv.Load(".env")
}
