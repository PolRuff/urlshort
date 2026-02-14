package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"
)

// Config holds application configuration
type Config struct {
	ServerAddr      string `env:"SERVER_ADDRESS" json:"server_address"`
	BaseURL         string `env:"BASE_URL" json:"base_url"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DatabaseDsn     string `env:"DATABASE_DSN" json:"database_dsn"`
	SecretKey       string `env:"SECRET_KEY" json:"secret_key"`
	AuditFile       string `env:"AUDIT_FILE" json:"audit_file"`
	AuditURL        string `env:"AUDIT_URL" json:"audit_url"`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS" json:"enable_https"`
}

// Load parses environment variables and command-line flags (from args) and returns application config.
// args should be like os.Args[1:].
// Priority: 1. Environment variables, 2. CLI flags (-a, -b, -f), 3. JSON config 4. Default values
func Load(args []string) (*Config, error) {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	var (
		serverAddr      = fs.String("a", "localhost:8080", "HTTP server address (e.g. localhost:8888)")
		baseURL         = fs.String("b", "http://localhost:8080", "Base URL for shortened links (e.g. http://localhost:8000)")
		fileStoragePath = fs.String("f", "./storage.json", "Path to the file storage (e.g. /path/to/storage.json)")
		databaseDsn     = fs.String("d", "", "Data source name (e.g. postgres://urlshort:urlshort@localhost:5432/urlshort?sslmode=disable)")
		secretKey       = fs.String("k", "supersecretkey", "Secret key for symmetrically sign cookie")
		auditFile       = fs.String("audit-file", "", "Path to the audit log file")
		auditURL        = fs.String("audit-url", "", "URL of the remote audit log server")
		enableHTTPS     = fs.Bool("s", false, "Enable HTTPS")
		configPathFlag  = fs.String("c", "", "Path to JSON config file")
		configPathAlt   = fs.String("config", "", "Path to JSON config file")
	)

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("failed to parse flags: %w", err)
	}

	// determine config path: flag > alt flag > env
	configPath := *configPathFlag
	if configPath == "" {
		configPath = *configPathAlt
	}
	if configPath == "" {
		configPath = os.Getenv("CONFIG")
	}

	// start with defaults (from flags)
	cfg := &Config{
		ServerAddr:      *serverAddr,
		BaseURL:         *baseURL,
		FileStoragePath: *fileStoragePath,
		DatabaseDsn:     *databaseDsn,
		SecretKey:       *secretKey,
		AuditFile:       *auditFile,
		AuditURL:        *auditURL,
		EnableHTTPS:     *enableHTTPS,
	}

	// collect explicitly set flags
	setFlags := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = true
	})

	// apply JSON config (lower priority than flags)
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		var jsonCfg Config
		if err := json.Unmarshal(data, &jsonCfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}

		if !setFlags["a"] && jsonCfg.ServerAddr != "" {
			cfg.ServerAddr = jsonCfg.ServerAddr
		}
		if !setFlags["b"] && jsonCfg.BaseURL != "" {
			cfg.BaseURL = jsonCfg.BaseURL
		}
		if !setFlags["f"] && jsonCfg.FileStoragePath != "" {
			cfg.FileStoragePath = jsonCfg.FileStoragePath
		}
		if !setFlags["d"] && jsonCfg.DatabaseDsn != "" {
			cfg.DatabaseDsn = jsonCfg.DatabaseDsn
		}
		if !setFlags["k"] && jsonCfg.SecretKey != "" {
			cfg.SecretKey = jsonCfg.SecretKey
		}
		if !setFlags["audit-file"] && jsonCfg.AuditFile != "" {
			cfg.AuditFile = jsonCfg.AuditFile
		}
		if !setFlags["audit-url"] && jsonCfg.AuditURL != "" {
			cfg.AuditURL = jsonCfg.AuditURL
		}
		if !setFlags["s"] {
			cfg.EnableHTTPS = jsonCfg.EnableHTTPS
		}
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
