package auth

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestTokenPath(t *testing.T) {
	configDir := "/home/user/.config/gagent-cli"

	readPath := TokenPath(configDir, ScopeRead)
	assert.Equal(t, "/home/user/.config/gagent-cli/token_read.json", readPath)

	writePath := TokenPath(configDir, ScopeWrite)
	assert.Equal(t, "/home/user/.config/gagent-cli/token_write.json", writePath)
}

func TestSaveAndLoadToken(t *testing.T) {
	// Create a temp directory for the test
	tempDir, err := os.MkdirTemp("", "gagent-cli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tokenPath := filepath.Join(tempDir, "token.json")

	// Create a test token
	token := &oauth2.Token{
		AccessToken:  "test-access-token",
		TokenType:    "Bearer",
		RefreshToken: "test-refresh-token",
		Expiry:       time.Now().Add(time.Hour),
	}

	// Save the token
	err = SaveToken(tokenPath, token)
	require.NoError(t, err)

	// Verify file permissions
	info, err := os.Stat(tokenPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	// Load the token
	loadedToken, err := LoadToken(tokenPath)
	require.NoError(t, err)
	assert.Equal(t, token.AccessToken, loadedToken.AccessToken)
	assert.Equal(t, token.TokenType, loadedToken.TokenType)
	assert.Equal(t, token.RefreshToken, loadedToken.RefreshToken)
}

func TestLoadToken_NotFound(t *testing.T) {
	token, err := LoadToken("/nonexistent/path/token.json")
	assert.Error(t, err)
	assert.Nil(t, token)
	assert.Contains(t, err.Error(), "token file not found")
}

func TestTokenExists(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gagent-cli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Initially no token
	assert.False(t, TokenExists(tempDir, ScopeRead))
	assert.False(t, TokenExists(tempDir, ScopeWrite))

	// Save a read token
	token := &oauth2.Token{AccessToken: "test"}
	tokenPath := TokenPath(tempDir, ScopeRead)
	err = SaveToken(tokenPath, token)
	require.NoError(t, err)

	// Now read exists, write still doesn't
	assert.True(t, TokenExists(tempDir, ScopeRead))
	assert.False(t, TokenExists(tempDir, ScopeWrite))
}

func TestDeleteToken(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gagent-cli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Save a token
	token := &oauth2.Token{AccessToken: "test"}
	tokenPath := TokenPath(tempDir, ScopeRead)
	err = SaveToken(tokenPath, token)
	require.NoError(t, err)

	// Verify it exists
	assert.True(t, TokenExists(tempDir, ScopeRead))

	// Delete it
	err = DeleteToken(tempDir, ScopeRead)
	require.NoError(t, err)

	// Verify it's gone
	assert.False(t, TokenExists(tempDir, ScopeRead))

	// Deleting again should not error
	err = DeleteToken(tempDir, ScopeRead)
	require.NoError(t, err)
}
