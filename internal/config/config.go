package config

import (
	"flag"
)

// Config holds the application configuration
type Config struct {
	ServerAddr string // Address to listen on HTTP-server (flag -a)
	BaseURL    string // Base URL for shortened links (flag -b)
}

// MustLoad parses command-line flags and returns a Config.
// Exits the program if required flags are missing or invalid.
func MustLoad() *Config {
	var (
		serverAddr = flag.String("a", "localhost:8080", "HTTP server address (e.g. localhost:8888)")
		baseURL    = flag.String("b", "http://localhost:8080", "Base URL for shortened links (e.g. http://localhost:8000)")
	)

	flag.Parse()

	return &Config{
		ServerAddr: *serverAddr,
		BaseURL:    *baseURL,
	}
}
