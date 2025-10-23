package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustLoad_PriorityEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("SERVER_ADDRESS", "0.0.0.0:9090")
	os.Setenv("BASE_URL", "http://env.com")

	// Ensure environment variables are cleared after the test
	defer func() {
		os.Unsetenv("SERVER_ADDRESS")
		os.Unsetenv("BASE_URL")
	}()

	// Call MustLoad with flags that should be overridden by env vars
	cfg := MustLoad([]string{"-a", "localhost:8080", "-b", "http://flag.com"})

	assert.Equal(t, "0.0.0.0:9090", cfg.ServerAddr) // env takes precedence
	assert.Equal(t, "http://env.com", cfg.BaseURL)  // env takes precedence
}

func TestMustLoad_PriorityFlag(t *testing.T) {
	// Ensure environment variables are not set
	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")

	// Call MustLoad with flags that override defaults
	cfg := MustLoad([]string{"-a", "127.0.0.1:8081", "-b", "http://flag.com"})

	assert.Equal(t, "127.0.0.1:8081", cfg.ServerAddr)
	assert.Equal(t, "http://flag.com", cfg.BaseURL)
}

func TestMustLoad_Defaults(t *testing.T) {
	// Ensure environment variables are not set
	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")

	// Call MustLoad with no arguments to use defaults
	cfg := MustLoad([]string{})

	assert.Equal(t, "localhost:8080", cfg.ServerAddr)
	assert.Equal(t, "http://localhost:8080", cfg.BaseURL)
}
