package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, "primary", cfg.DefaultCalendar)
	assert.Equal(t, "json", cfg.OutputFormat)
	assert.Equal(t, 30, cfg.TimeoutSeconds)
	assert.False(t, cfg.AuditLog)
}

func TestSaveAndLoad(t *testing.T) {
	// Override config dir for test
	tempDir, err := os.MkdirTemp("", "gagent-cli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Save config dir
	oldConfigDir := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldConfigDir)

	// Create and save a config
	cfg := &Config{
		ClientID:        "test-client-id.apps.googleusercontent.com",
		ClientSecret:    "test-client-secret",
		DefaultCalendar: "custom-calendar",
		OutputFormat:    "json",
		TimeoutSeconds:  60,
		AuditLog:        true,
	}

	err = Save(cfg)
	require.NoError(t, err)

	// Verify file permissions
	configPath := filepath.Join(tempDir, DefaultConfigDirName, ConfigFileName)
	info, err := os.Stat(configPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	// Load and verify
	loadedCfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, cfg.ClientID, loadedCfg.ClientID)
	assert.Equal(t, cfg.ClientSecret, loadedCfg.ClientSecret)
	assert.Equal(t, cfg.DefaultCalendar, loadedCfg.DefaultCalendar)
	assert.Equal(t, cfg.TimeoutSeconds, loadedCfg.TimeoutSeconds)
	assert.Equal(t, cfg.AuditLog, loadedCfg.AuditLog)
}

func TestLoad_NotFound(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gagent-cli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	oldConfigDir := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldConfigDir)

	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "gagent-cli auth setup")
}

func TestExists(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gagent-cli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	oldConfigDir := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldConfigDir)

	// Initially doesn't exist
	assert.False(t, Exists())

	// Save a config
	cfg := DefaultConfig()
	cfg.ClientID = "test"
	cfg.ClientSecret = "test"
	err = Save(cfg)
	require.NoError(t, err)

	// Now it exists
	assert.True(t, Exists())
}

func TestSetAndGet(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gagent-cli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	oldConfigDir := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldConfigDir)

	// First save a base config
	cfg := &Config{
		ClientID:        "test-id",
		ClientSecret:    "test-secret",
		DefaultCalendar: "primary",
	}
	err = Save(cfg)
	require.NoError(t, err)

	// Test Set
	err = Set("default_calendar", "work")
	require.NoError(t, err)

	err = Set("audit_log", "true")
	require.NoError(t, err)

	// Test Get
	value, err := Get("default_calendar")
	require.NoError(t, err)
	assert.Equal(t, "work", value)

	value, err = Get("audit_log")
	require.NoError(t, err)
	assert.Equal(t, "true", value)

	// Test invalid key
	err = Set("invalid_key", "value")
	assert.Error(t, err)

	_, err = Get("invalid_key")
	assert.Error(t, err)
}
