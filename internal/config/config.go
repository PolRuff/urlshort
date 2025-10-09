package config

import "flag"

// Config holds application configuration
type Config struct {
	ServerAddr string // server address to listen on (e.g. localhost:8080)
	BaseURL    string // base URL for shortened links (e.g. http://localhost:8080)
}

// MustLoad parses command-line flags and returns application config
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
