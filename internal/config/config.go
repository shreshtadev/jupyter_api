package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Environment int

// Define the enum values using iota
const (
	Local Environment = iota
	Dev
	Prod
)

// String method allows the enum to be printed as text
func (e Environment) String() string {
	return [...]string{"local", "dev", "prod"}[e]
}

type Config struct {
	Addr    string // HTTP address
	DSN     string // MySQL/MariaDB DSN
	APP_ENV string // local,dev,prod
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

	app_env := os.Getenv("APP_ENV")
	if app_env == "" {
		app_env = Local.String()
	}

	return &Config{
		Addr:    addr,
		DSN:     dsn,
		APP_ENV: app_env,
	}
}

func LoadEnv() {
	_ = godotenv.Load(".env")
}
