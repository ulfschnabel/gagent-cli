package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetScopes_Read(t *testing.T) {
	scopes := GetScopes(ScopeRead)
	assert.NotEmpty(t, scopes)
	assert.Contains(t, scopes, "https://www.googleapis.com/auth/gmail.readonly")
	assert.Contains(t, scopes, "https://www.googleapis.com/auth/calendar.readonly")
	assert.Contains(t, scopes, "https://www.googleapis.com/auth/documents.readonly")
	assert.Contains(t, scopes, "https://www.googleapis.com/auth/spreadsheets.readonly")
	assert.Contains(t, scopes, "https://www.googleapis.com/auth/presentations.readonly")
	assert.Contains(t, scopes, "https://www.googleapis.com/auth/drive.readonly")
}

func TestGetScopes_Write(t *testing.T) {
	scopes := GetScopes(ScopeWrite)
	assert.NotEmpty(t, scopes)
	assert.Contains(t, scopes, "https://www.googleapis.com/auth/gmail.send")
	assert.Contains(t, scopes, "https://www.googleapis.com/auth/gmail.modify")
	assert.Contains(t, scopes, "https://www.googleapis.com/auth/calendar")
	assert.Contains(t, scopes, "https://www.googleapis.com/auth/documents")
	assert.Contains(t, scopes, "https://www.googleapis.com/auth/spreadsheets")
	assert.Contains(t, scopes, "https://www.googleapis.com/auth/presentations")
	assert.Contains(t, scopes, "https://www.googleapis.com/auth/drive.file")
}

func TestGetScopes_Invalid(t *testing.T) {
	scopes := GetScopes("invalid")
	assert.Nil(t, scopes)
}

func TestIsValidScopeType(t *testing.T) {
	tests := []struct {
		name     string
		scope    string
		expected bool
	}{
		{"read is valid", "read", true},
		{"write is valid", "write", true},
		{"empty is invalid", "", false},
		{"other is invalid", "other", false},
		{"READ uppercase is invalid", "READ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidScopeType(tt.scope)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTokenFilename(t *testing.T) {
	tests := []struct {
		scopeType ScopeType
		expected  string
	}{
		{ScopeRead, "token_read.json"},
		{ScopeWrite, "token_write.json"},
		{"invalid", ""},
	}

	for _, tt := range tests {
		t.Run(string(tt.scopeType), func(t *testing.T) {
			result := TokenFilename(tt.scopeType)
			assert.Equal(t, tt.expected, result)
		})
	}
}
