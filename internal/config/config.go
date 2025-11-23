package config

import (
	"flag"

	"github.com/rs/zerolog/log"

	"github.com/caarlos0/env/v6"
)

// Config holds application configuration
type Config struct {
	ServerAddr      string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDsn     string `env:"DATABASE_DSN"`
	SecretKey       string `env:"SECRET_KEY"`
}

// MustLoad parses environment variables and command-line flags (from args) and returns application config.
// args should be like os.Args[1:].
// Priority: 1. Environment variables, 2. CLI flags (-a, -b, -f), 3. Default values
// Panics on fatal errors (e.g., missing required flags/env vars, parse errors).
func MustLoad(args []string) *Config {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	var (
		serverAddr      = fs.String("a", "localhost:8080", "HTTP server address (e.g. localhost:8888)")
		baseURL         = fs.String("b", "http://localhost:8080", "Base URL for shortened links (e.g. http://localhost:8000)")
		fileStoragePath = fs.String("f", "/tmp/storage.json", "Path to the file storage (e.g. /path/to/storage.json)")
		databaseDsn     = fs.String("d", "", "Data source name (e.g. postgres://urlshort:urlshort@localhost:5432/urlshort?sslmode=disable)")
		secretKey       = fs.String("s", "supersecretkey", "Secret key for symmetrically sign cookie")
	)

	err := fs.Parse(args)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse flags")
	}

	cfg := &Config{
		ServerAddr:      *serverAddr,
		BaseURL:         *baseURL,
		FileStoragePath: *fileStoragePath,
		DatabaseDsn:     *databaseDsn,
		SecretKey:       *secretKey,
	}

	// Load configuration from environment variables (they take precedence)
	err = env.Parse(cfg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse config from environment")
	}

	if cfg.ServerAddr == "" {
		log.Error().Msg("required flag -a or env SERVER_ADDRESS is missing")
	}
	if cfg.BaseURL == "" {
		log.Error().Msg("required flag -b or env BASE_URL is missing")
	}
	if cfg.SecretKey == "supersecretkey" {
		log.Warn().Msg("used default secret key")
	}

	return cfg
}
