package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
)

const (
	// TokenReadFile is the filename for the read-only OAuth token.
	TokenReadFile = "token_read.json"
	// TokenWriteFile is the filename for the write OAuth token.
	TokenWriteFile = "token_write.json"
)

// TokenFilename returns the token filename for the given scope type.
func TokenFilename(scopeType ScopeType) string {
	switch scopeType {
	case ScopeRead:
		return TokenReadFile
	case ScopeWrite:
		return TokenWriteFile
	default:
		return ""
	}
}

// TokenPath returns the full path to the token file for the given scope type.
func TokenPath(configDir string, scopeType ScopeType) string {
	return filepath.Join(configDir, TokenFilename(scopeType))
}

// SaveToken saves an OAuth token to the specified file with 0600 permissions.
func SaveToken(path string, token *oauth2.Token) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create token file: %w", err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(token); err != nil {
		return fmt.Errorf("failed to encode token: %w", err)
	}

	return nil
}

// LoadToken loads an OAuth token from the specified file.
func LoadToken(path string) (*oauth2.Token, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("token file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to open token file: %w", err)
	}
	defer f.Close()

	var token oauth2.Token
	if err := json.NewDecoder(f).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to decode token: %w", err)
	}

	return &token, nil
}

// TokenExists checks if a token file exists for the given scope type.
func TokenExists(configDir string, scopeType ScopeType) bool {
	path := TokenPath(configDir, scopeType)
	_, err := os.Stat(path)
	return err == nil
}

// DeleteToken removes the token file for the given scope type.
func DeleteToken(configDir string, scopeType ScopeType) error {
	path := TokenPath(configDir, scopeType)
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return nil // Token doesn't exist, nothing to delete
		}
		return fmt.Errorf("failed to delete token: %w", err)
	}
	return nil
}
