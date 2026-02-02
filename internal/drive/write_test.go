package drive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShareOptions(t *testing.T) {
	opts := ShareOptions{
		FileID:       "file-to-share",
		EmailAddress: "user@example.com",
		Role:         "writer",
		Type:         "user",
		SendEmail:    true,
	}

	assert.Equal(t, "file-to-share", opts.FileID)
	assert.Equal(t, "user@example.com", opts.EmailAddress)
	assert.Equal(t, "writer", opts.Role)
	assert.Equal(t, "user", opts.Type)
	assert.True(t, opts.SendEmail)
}

func TestShareOptionsDefaults(t *testing.T) {
	// Test that empty role/type get handled correctly by Share method
	opts := ShareOptions{
		FileID:       "file-123",
		EmailAddress: "reader@example.com",
	}

	// Verify defaults aren't set in the struct itself
	assert.Equal(t, "", opts.Role)
	assert.Equal(t, "", opts.Type)
	assert.False(t, opts.SendEmail)
}

func TestShareOptionsRoles(t *testing.T) {
	validRoles := []string{"reader", "writer", "commenter", "owner"}

	for _, role := range validRoles {
		opts := ShareOptions{
			FileID:       "file-123",
			EmailAddress: "test@example.com",
			Role:         role,
			Type:         "user",
		}
		assert.Equal(t, role, opts.Role)
	}
}

func TestShareOptionsTypes(t *testing.T) {
	validTypes := []string{"user", "group", "domain", "anyone"}

	for _, permType := range validTypes {
		opts := ShareOptions{
			FileID:       "file-123",
			EmailAddress: "test@example.com",
			Role:         "reader",
			Type:         permType,
		}
		assert.Equal(t, permType, opts.Type)
	}
}
