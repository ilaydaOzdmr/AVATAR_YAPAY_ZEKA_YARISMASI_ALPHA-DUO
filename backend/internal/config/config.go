package config

import "os"

// Config holds the runtime configuration loaded from environment variables.
type Config struct {
	// DSN is the MySQL connection string.
	DSN string
	// Addr is the HTTP listen address.
	Addr string
}

// Load reads configuration from environment variables, applying defaults.
func Load() Config {
	return Config{
		DSN:  getenv("DATABASE_DSN", "avatar:avatar@tcp(localhost:3306)/avatar?charset=utf8mb4"),
		Addr: getenv("HTTP_ADDR", ":8080"),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
