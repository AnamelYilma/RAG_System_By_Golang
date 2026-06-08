package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds app-level settings used by the HTTP server and document indexer.
type Config struct {
	Port      string
	PDFDir    string
	ChunkSize int
	Overlap   int
}

// Load reads environment variables and applies safe defaults.
func Load() Config {
	return Config{
		Port:      envString("APP_PORT", "8080"),
		PDFDir:    envString("PDF_DIR", "./PDF"),
		ChunkSize: envInt("CHUNK_SIZE", 450),
		Overlap:   envInt("CHUNK_OVERLAP", 90),
	}
}

// Addr returns the HTTP listen address for the server.
func (c Config) Addr() string {
	return ":" + strings.TrimPrefix(c.Port, ":")
}

func envString(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return fallback
	}

	return value
}
