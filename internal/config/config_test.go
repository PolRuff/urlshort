package config

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_PriorityEnv(t *testing.T) {
	// Set environment variables
	t.Setenv("SERVER_ADDRESS", "0.0.0.0:9090")
	t.Setenv("BASE_URL", "http://env.com")
	t.Setenv("FILE_STORAGE_PATH", "/env/path/storage.json")
	t.Setenv("DATABASE_DSN", "postgres://music:passworduser@localhost:2345/music?sslmode=enable")
	t.Setenv("SECRET_KEY", "verysupersecretkey")
	t.Setenv("AUDIT_FILE", "/var/log/audit.log")
	t.Setenv("AUDIT_URL", "https://audit.example.com/logs")
	t.Setenv("ENABLE_HTTPS", "true")

	// Call Load with flags that should be overridden by env vars
	cfg, err := Load([]string{
		"-a", "localhost:8080",
		"-b", "http://flag.com",
		"-f", "/flag/path/storage.json",
		"-d", "postgres://picture:passworduserpicture@localhost:4523/picture?sslmode=default",
		"-k", "veryverysupersecretkey",
		"--audit-file", "/tmp/flag_audit.log",
		"--audit-url", "http://flag-audit.local",
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
	// Call Load with flags that override defaults
	cfg, err := Load([]string{
		"-a", "127.0.0.1:8081",
		"-b", "http://flag.com",
		"-f", "/flag/path/storage.json",
		"-d", "postgres://picture:passworduserpicture@localhost:4523/picture?sslmode=default",
		"-k", "verysupersecretkey",
		"--audit-file", "/tmp/flag_audit.log",
		"--audit-url", "http://flag-audit.local",
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

func TestLoad_Help(t *testing.T) {
	_, err := Load([]string{"--help"})

	require.ErrorIs(t, err, flag.ErrHelp)
}

func TestLoad_InvalidFlags(t *testing.T) {
	_, err := Load([]string{"--invalid flag"})

	require.Error(t, err)
}

func TestLoad_Defaults(t *testing.T) {
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

func TestLoad_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	jsonContent := `{
		"server_address": "json:8080",
		"base_url": "http://json",
		"file_storage_path": "/json/storage.json",
		"database_dsn": "postgres://json",
		"secret_key": "verysupersecretkey",
		"audit_file": "/tmp/flag_audit.log",
		"audit_url": "http://flag-audit.local",
		"enable_https": true
	}`

	require.NoError(t, os.WriteFile(configFile, []byte(jsonContent), 0644))

	cfg, err := Load([]string{"-c", configFile})
	require.NoError(t, err)

	assert.Equal(t, "json:8080", cfg.ServerAddr)
	assert.Equal(t, "http://json", cfg.BaseURL)
	assert.Equal(t, "/json/storage.json", cfg.FileStoragePath)
	assert.Equal(t, "postgres://json", cfg.DatabaseDsn)
	assert.Equal(t, "verysupersecretkey", cfg.SecretKey)
	assert.Equal(t, "/tmp/flag_audit.log", cfg.AuditFile)
	assert.Equal(t, "http://flag-audit.local", cfg.AuditURL)
	assert.True(t, cfg.EnableHTTPS)
}

func TestLoad_InvalidConfigPath(t *testing.T) {
	_, err := Load([]string{"-c", "invalid path to config file"})
	require.Error(t, err)
}

func TestLoad_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	invalidConfigFile := filepath.Join(tmpDir, "invalid_config.json")

	invalidJSONContent := `invalid json content`

	require.NoError(t, os.WriteFile(invalidConfigFile, []byte(invalidJSONContent), 0644))

	_, err := Load([]string{"-c", invalidConfigFile})
	require.Error(t, err)
}
