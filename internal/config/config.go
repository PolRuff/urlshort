package config

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config holds application configuration
type Config struct {
	ServerAddr      string `mapstructure:"server_address"`
	BaseURL         string `mapstructure:"base_url"`
	FileStoragePath string `mapstructure:"file_storage_path"`
	DatabaseDsn     string `mapstructure:"database_dsn"`
	SecretKey       string `mapstructure:"secret_key"`
	AuditFile       string `mapstructure:"audit_file"`
	AuditURL        string `mapstructure:"audit_url"`
	EnableHTTPS     bool   `mapstructure:"enable_https"`
}

// Load parses environment variables and command-line flags (from args) and returns application config.
// args should be like os.Args[1:].
// Priority: 1. Environment variables, 2. CLI flags (-a, -b, -f), 3. JSON config 4. Default values
func Load(args []string) (*Config, error) {
	v := viper.New()

	// Environment variables
	v.AutomaticEnv()

	// By default, command-line flags is higher priority than environment variables.
	// Change it by viper.Set()
	keys := [...]string{
		"server_address",
		"base_url",
		"file_storage_path",
		"database_dsn",
		"secret_key",
		"audit_file",
		"audit_url",
		"enable_https",
	}
	for _, key := range keys {
		if v.IsSet(key) {
			v.Set(key, v.Get(key))
		}
	}

	// Command-line flags with default values
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)

	fs.StringP("server_address", "a", "localhost:8080", "HTTP server address")
	fs.StringP("base_url", "b", "http://localhost:8080", "Base URL for shortened links")
	fs.StringP("file_storage_path", "f", "./storage.json", "File storage path")
	fs.StringP("database_dsn", "d", "", "Database DSN")
	fs.StringP("secret_key", "k", "supersecretkey", "Secret key")
	fs.String("audit_file", "", "Audit log file path")
	fs.String("audit_url", "", "Audit log remote URL")
	fs.BoolP("enable_https", "s", false, "Enable HTTPS")
	fs.StringP("config", "c", "", "Path to config file")

	fs.SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		name = strings.ReplaceAll(name, "-", "_")
		return pflag.NormalizedName(name)
	})

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return nil, flag.ErrHelp
		}
		return nil, fmt.Errorf("failed to parse flags: %w", err)
	}

	if err := v.BindPFlags(fs); err != nil {
		return nil, fmt.Errorf("failed to bind flags: %w", err)
	}

	// Configuration file
	configPath := v.GetString("config")
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
