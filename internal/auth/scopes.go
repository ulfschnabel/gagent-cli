// Package auth handles OAuth authentication with Google APIs.
package auth

// ScopeType represents the type of OAuth scope (read or write).
type ScopeType string

const (
	// ScopeRead represents read-only access scopes.
	ScopeRead ScopeType = "read"
	// ScopeWrite represents write access scopes.
	ScopeWrite ScopeType = "write"
)

// ReadScopes defines the low-risk read-only OAuth scopes.
var ReadScopes = []string{
	"https://www.googleapis.com/auth/gmail.readonly",
	"https://www.googleapis.com/auth/calendar.readonly",
	"https://www.googleapis.com/auth/documents.readonly",
	"https://www.googleapis.com/auth/spreadsheets.readonly",
	"https://www.googleapis.com/auth/presentations.readonly",
	"https://www.googleapis.com/auth/drive.readonly",
	"https://www.googleapis.com/auth/contacts.readonly",
}

// WriteScopes defines the high-risk write access OAuth scopes.
var WriteScopes = []string{
	"https://www.googleapis.com/auth/gmail.send",
	"https://www.googleapis.com/auth/gmail.modify",
	"https://www.googleapis.com/auth/calendar",
	"https://www.googleapis.com/auth/documents",
	"https://www.googleapis.com/auth/spreadsheets",
	"https://www.googleapis.com/auth/presentations",
	"https://www.googleapis.com/auth/drive.file",
	"https://www.googleapis.com/auth/contacts",
}

// GetScopes returns the OAuth scopes for the given scope type.
func GetScopes(scopeType ScopeType) []string {
	switch scopeType {
	case ScopeRead:
		return ReadScopes
	case ScopeWrite:
		return WriteScopes
	default:
		return nil
	}
}

// IsValidScopeType checks if the given scope type is valid.
func IsValidScopeType(s string) bool {
	return s == string(ScopeRead) || s == string(ScopeWrite)
}
