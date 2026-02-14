package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_PriorityEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("SERVER_ADDRESS", "0.0.0.0:9090")
	os.Setenv("BASE_URL", "http://env.com")
	os.Setenv("FILE_STORAGE_PATH", "/env/path/storage.json")
	os.Setenv("DATABASE_DSN", "postgres://music:passworduser@localhost:2345/music?sslmode=enable")
	os.Setenv("SECRET_KEY", "verysupersecretkey")
	os.Setenv("AUDIT_FILE", "/var/log/audit.log")
	os.Setenv("AUDIT_URL", "https://audit.example.com/logs")
	os.Setenv("ENABLE_HTTPS", "true")

	// Ensure environment variables are cleared after the test
	defer func() {
		os.Unsetenv("SERVER_ADDRESS")
		os.Unsetenv("BASE_URL")
		os.Unsetenv("FILE_STORAGE_PATH")
		os.Unsetenv("DATABASE_DSN")
		os.Unsetenv("SECRET_KEY")
		os.Unsetenv("AUDIT_FILE")
		os.Unsetenv("AUDIT_URL")
		os.Unsetenv("ENABLE_HTTPS")
	}()

	// Call Load with flags that should be overridden by env vars
	cfg, err := Load([]string{
		"-a", "localhost:8080",
		"-b", "http://flag.com",
		"-f", "/flag/path/storage.json",
		"-d", "postgres://picture:passworduserpicture@localhost:4523/picture?sslmode=default",
		"-k", "veryverysupersecretkey",
		"-audit-file", "/tmp/flag_audit.log",
		"-audit-url", "http://flag-audit.local",
		"-s",
	})
	require.NoError(t, err, "Load should not return an error")

	assert.Equal(t, "0.0.0.0:9090", cfg.ServerAddr)                                                       // env takes precedence
	assert.Equal(t, "http://env.com", cfg.BaseURL)                                                        // env takes precedence
	assert.Equal(t, "/env/path/storage.json", cfg.FileStoragePath)                                        // env takes precedence
	assert.Equal(t, "postgres://music:passworduser@localhost:2345/music?sslmode=enable", cfg.DatabaseDsn) // env takes precedence
	assert.Equal(t, "verysupersecretkey", cfg.SecretKey)                                                  // env takes precedence
	assert.Equal(t, "/var/log/audit.log", cfg.AuditFile)                                                  // env takes precedence
	assert.Equal(t, "https://audit.example.com/logs", cfg.AuditURL)                                       // env takes precedence
	assert.True(t, cfg.EnableHTTPS)                                                                       // env takes precedence
}

func TestLoad_PriorityFlag(t *testing.T) {
	// Ensure environment variables are not set
	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")
	os.Unsetenv("FILE_STORAGE_PATH")
	os.Unsetenv("DATABASE_DSN")
	os.Unsetenv("SECRET_KEY")
	os.Unsetenv("AUDIT_FILE")
	os.Unsetenv("AUDIT_URL")
	os.Unsetenv("ENABLE_HTTPS")

	// Call Load with flags that override defaults
	cfg, err := Load([]string{
		"-a", "127.0.0.1:8081",
		"-b", "http://flag.com",
		"-f", "/flag/path/storage.json",
		"-d", "postgres://picture:passworduserpicture@localhost:4523/picture?sslmode=default",
		"-k", "verysupersecretkey",
		"-audit-file", "/tmp/flag_audit.log",
		"-audit-url", "http://flag-audit.local",
		"-s",
	})
	require.NoError(t, err, "Load should not return an error")

	assert.Equal(t, "127.0.0.1:8081", cfg.ServerAddr)
	assert.Equal(t, "http://flag.com", cfg.BaseURL)
	assert.Equal(t, "/flag/path/storage.json", cfg.FileStoragePath)
	assert.Equal(t, "postgres://picture:passworduserpicture@localhost:4523/picture?sslmode=default", cfg.DatabaseDsn)
	assert.Equal(t, "verysupersecretkey", cfg.SecretKey)
	assert.Equal(t, "/tmp/flag_audit.log", cfg.AuditFile)
	assert.Equal(t, "http://flag-audit.local", cfg.AuditURL)
	assert.True(t, cfg.EnableHTTPS)
}

func TestLoad_Defaults(t *testing.T) {
	// Ensure environment variables are not set
	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")
	os.Unsetenv("FILE_STORAGE_PATH")
	os.Unsetenv("DATABASE_DSN")
	os.Unsetenv("SECRET_KEY")
	os.Unsetenv("AUDIT_FILE")
	os.Unsetenv("AUDIT_URL")
	os.Unsetenv("ENABLE_HTTPS")

	// Call Load with no arguments to use defaults
	cfg, err := Load([]string{})
	require.NoError(t, err, "Load should not return an error")

	assert.Equal(t, "localhost:8080", cfg.ServerAddr)
	assert.Equal(t, "http://localhost:8080", cfg.BaseURL)
	assert.Equal(t, "./storage.json", cfg.FileStoragePath)
	assert.Equal(t, "", cfg.DatabaseDsn)
	assert.Equal(t, "supersecretkey", cfg.SecretKey)
	assert.Equal(t, "", cfg.AuditFile)
	assert.Equal(t, "", cfg.AuditURL)
	assert.False(t, cfg.EnableHTTPS)
}
