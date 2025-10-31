package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

// Config holds application configuration
type Config struct {
	ServerAddr      string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

// MustLoad parses environment variables and command-line flags (from args) and returns application config.
// args should be like os.Args[1:].
// Priority: 1. Environment variables, 2. CLI flags (-a, -b, -f), 3. Default values
func MustLoad(args []string) *Config {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	var (
		serverAddr      = fs.String("a", "localhost:8080", "HTTP server address (e.g. localhost:8888)")
		baseURL         = fs.String("b", "http://localhost:8080", "Base URL for shortened links (e.g. http://localhost:8000)")
		fileStoragePath = fs.String("f", "./storage.json", "Path to the file storage (e.g. /path/to/storage.json)")
	)

	err := fs.Parse(args)
	if err != nil {
		log.Printf("Failed to parse flags: %v", err)
	}

	cfg := &Config{
		ServerAddr:      *serverAddr,
		BaseURL:         *baseURL,
		FileStoragePath: *fileStoragePath,
	}

	// Load configuration from environment variables (they take precedence)
	err = env.Parse(cfg)
	if err != nil {
		log.Printf("Failed to parse config from environment: %v. Using flags or defaults.", err)
	}

	return cfg
}
