package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

// Config holds application configuration
type Config struct {
	ServerAddr string `env:"SERVER_ADDRESS"`
	BaseURL    string `env:"BASE_URL"`
}

// MustLoad parses environment variables and command-line flags (from args) and returns application config.
// args should be like os.Args[1:].
// Priority: 1. Environment variables, 2. CLI flags (-a, -b), 3. Default values
func MustLoad(args []string) *Config {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	var (
		serverAddr = fs.String("a", "localhost:8080", "HTTP server address (e.g. localhost:8888)")
		baseURL    = fs.String("b", "http://localhost:8080", "Base URL for shortened links (e.g. http://localhost:8000)")
	)

	err := fs.Parse(args)
	if err != nil {
		log.Printf("Failed to parse flags: %v", err)
	}

	cfg := &Config{
		ServerAddr: *serverAddr,
		BaseURL:    *baseURL,
	}

	// Load configuration from environment variables (they take precedence)
	err = env.Parse(cfg)
	if err != nil {
		log.Printf("Failed to parse config from environment: %v. Using flags or defaults.", err)
	}

	return cfg
}
