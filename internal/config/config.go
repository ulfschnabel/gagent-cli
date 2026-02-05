// Package config handles configuration management for gagent-cli.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// ConfigFileName is the name of the configuration file.
	ConfigFileName = "config.json"
	// DefaultConfigDirName is the default configuration directory name.
	DefaultConfigDirName = "gagent-cli"
)

// Config holds the application configuration.
type Config struct {
	ClientID        string `json:"client_id"`
	ClientSecret    string `json:"client_secret"`
	RedirectURL     string `json:"redirect_url,omitempty"`
	DefaultCalendar string `json:"default_calendar,omitempty"`
	OutputFormat    string `json:"output_format,omitempty"`
	TimeoutSeconds  int    `json:"timeout_seconds,omitempty"`
	AuditLog        bool   `json:"audit_log,omitempty"`
}

// DefaultConfig returns a configuration with default values.
func DefaultConfig() *Config {
	return &Config{
		DefaultCalendar: "primary",
		OutputFormat:    "json",
		TimeoutSeconds:  30,
		AuditLog:        false,
	}
}

// GetConfigDir returns the configuration directory path.
// It uses XDG_CONFIG_HOME if set, otherwise ~/.config/gagent-cli.
func GetConfigDir() (string, error) {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		configHome = filepath.Join(homeDir, ".config")
	}
	return filepath.Join(configHome, DefaultConfigDirName), nil
}

// GetConfigPath returns the full path to the configuration file.
func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, ConfigFileName), nil
}

// EnsureConfigDir creates the configuration directory if it doesn't exist.
func EnsureConfigDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return configDir, nil
}

// Load reads the configuration from the config file.
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("configuration not found. Run: gagent-cli auth setup")
		}
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	config := DefaultConfig()
	if err := json.NewDecoder(f).Decode(config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// Save writes the configuration to the config file with 0600 permissions.
func Save(config *Config) error {
	configDir, err := EnsureConfigDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(configDir, ConfigFileName)
	f, err := os.OpenFile(configPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Exists checks if the configuration file exists.
func Exists() bool {
	configPath, err := GetConfigPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(configPath)
	return err == nil
}

// Set updates a specific configuration value.
func Set(key, value string) error {
	config, err := Load()
	if err != nil {
		// If config doesn't exist, create a new one
		config = DefaultConfig()
	}

	switch key {
	case "redirect_url":
		config.RedirectURL = value
	case "default_calendar":
		config.DefaultCalendar = value
	case "output_format":
		if value != "json" {
			return fmt.Errorf("unsupported output format: %s (only 'json' is supported)", value)
		}
		config.OutputFormat = value
	case "audit_log":
		config.AuditLog = value == "true"
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	return Save(config)
}

// Get retrieves a specific configuration value.
func Get(key string) (string, error) {
	config, err := Load()
	if err != nil {
		return "", err
	}

	switch key {
	case "client_id":
		return config.ClientID, nil
	case "client_secret":
		return config.ClientSecret, nil
	case "redirect_url":
		return config.RedirectURL, nil
	case "default_calendar":
		return config.DefaultCalendar, nil
	case "output_format":
		return config.OutputFormat, nil
	case "audit_log":
		if config.AuditLog {
			return "true", nil
		}
		return "false", nil
	default:
		return "", fmt.Errorf("unknown configuration key: %s", key)
	}
}
