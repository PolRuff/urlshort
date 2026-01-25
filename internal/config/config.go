package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

// Config holds application configuration
type Config struct {
	ServerAddr      string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDsn     string `env:"DATABASE_DSN"`
	SecretKey       string `env:"SECRET_KEY"`
	AuditFile       string `env:"AUDIT_FILE"`
	AuditURL        string `env:"AUDIT_URL"`
}

// Load parses environment variables and command-line flags (from args) and returns application config.
// args should be like os.Args[1:].
// Priority: 1. Environment variables, 2. CLI flags (-a, -b, -f), 3. Default values
func Load(args []string) (*Config, error) {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	var (
		serverAddr      = fs.String("a", "localhost:8080", "HTTP server address (e.g. localhost:8888)")
		baseURL         = fs.String("b", "http://localhost:8080", "Base URL for shortened links (e.g. http://localhost:8000)")
		fileStoragePath = fs.String("f", "./storage.json", "Path to the file storage (e.g. /path/to/storage.json)")
		databaseDsn     = fs.String("d", "", "Data source name (e.g. postgres://urlshort:urlshort@localhost:5432/urlshort?sslmode=disable)")
		secretKey       = fs.String("s", "supersecretkey", "Secret key for symmetrically sign cookie")
		auditFile       = fs.String("audit-file", "", "Path to the audit log file")
		auditURL        = fs.String("audit-url", "", "URL of the remote audit log server")
	)

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("failed to parse flags: %w", err)
	}

	cfg := &Config{
		ServerAddr:      *serverAddr,
		BaseURL:         *baseURL,
		FileStoragePath: *fileStoragePath,
		DatabaseDsn:     *databaseDsn,
		SecretKey:       *secretKey,
		AuditFile:       *auditFile,
		AuditURL:        *auditURL,
	}

	// Load configuration from environment variables (they take precedence)
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config from environment: %w", err)
	}

	if cfg.ServerAddr == "" {
		return nil, fmt.Errorf("required flag -a or env SERVER_ADDRESS is missing")
	}
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("required flag -b or env BASE_URL is missing")
	}

	return cfg, nil
}
