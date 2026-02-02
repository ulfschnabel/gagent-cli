package drive

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/api/drive/v3"
)

func TestListOptions(t *testing.T) {
	opts := ListOptions{
		Query:      "name contains 'test'",
		MimeType:   MimeTypeDocument,
		ParentID:   "parent-folder-id",
		MaxResults: 50,
		PageToken:  "next-page-token",
		Trashed:    false,
		OrderBy:    "name",
	}

	assert.Equal(t, "name contains 'test'", opts.Query)
	assert.Equal(t, MimeTypeDocument, opts.MimeType)
	assert.Equal(t, "parent-folder-id", opts.ParentID)
	assert.Equal(t, int64(50), opts.MaxResults)
	assert.Equal(t, "next-page-token", opts.PageToken)
	assert.False(t, opts.Trashed)
	assert.Equal(t, "name", opts.OrderBy)
}

func TestParseFileToSummary(t *testing.T) {
	driveFile := &drive.File{
		Id:           "file-123",
		Name:         "test-document.docx",
		MimeType:     MimeTypeDocument,
		Size:         4096,
		ModifiedTime: "2024-01-15T10:30:00Z",
		Parents:      []string{"parent-1", "parent-2"},
		WebViewLink:  "https://docs.google.com/document/d/file-123",
		Starred:      true,
		Trashed:      false,
	}

	summary := parseFileToSummary(driveFile)

	assert.Equal(t, "file-123", summary.ID)
	assert.Equal(t, "test-document.docx", summary.Name)
	assert.Equal(t, MimeTypeDocument, summary.MimeType)
	assert.Equal(t, int64(4096), summary.Size)
	assert.Equal(t, "2024-01-15T10:30:00Z", summary.ModifiedTime)
	assert.Len(t, summary.Parents, 2)
	assert.Equal(t, "https://docs.google.com/document/d/file-123", summary.WebViewLink)
	assert.True(t, summary.Starred)
	assert.False(t, summary.Trashed)
}

func TestParseFileToSummary_EmptyParents(t *testing.T) {
	driveFile := &drive.File{
		Id:       "file-456",
		Name:     "orphan-file.txt",
		MimeType: "text/plain",
	}

	summary := parseFileToSummary(driveFile)

	assert.Equal(t, "file-456", summary.ID)
	assert.Nil(t, summary.Parents)
}

func TestParseFileToMetadata(t *testing.T) {
	driveFile := &drive.File{
		Id:             "file-789",
		Name:           "detailed-file.pdf",
		MimeType:       "application/pdf",
		Size:           8192,
		CreatedTime:    "2024-01-10T08:00:00Z",
		ModifiedTime:   "2024-01-15T10:30:00Z",
		Parents:        []string{"parent-folder"},
		WebViewLink:    "https://drive.google.com/file/d/file-789",
		WebContentLink: "https://drive.google.com/uc?id=file-789",
		IconLink:       "https://drive.google.com/icon.png",
		Starred:        false,
		Trashed:        false,
		Shared:         true,
		Description:    "A detailed PDF document",
		Properties:     map[string]string{"custom": "value"},
		Owners: []*drive.User{
			{DisplayName: "Owner Name", EmailAddress: "owner@example.com", PhotoLink: "https://photo.url"},
		},
		LastModifyingUser: &drive.User{
			DisplayName:  "Last Modifier",
			EmailAddress: "modifier@example.com",
		},
	}

	meta := parseFileToMetadata(driveFile)

	assert.Equal(t, "file-789", meta.ID)
	assert.Equal(t, "detailed-file.pdf", meta.Name)
	assert.Equal(t, "application/pdf", meta.MimeType)
	assert.Equal(t, int64(8192), meta.Size)
	assert.Equal(t, "2024-01-10T08:00:00Z", meta.CreatedTime)
	assert.Equal(t, "2024-01-15T10:30:00Z", meta.ModifiedTime)
	assert.True(t, meta.Shared)
	assert.Equal(t, "A detailed PDF document", meta.Description)
	assert.Equal(t, "value", meta.Properties["custom"])

	// Check owners
	assert.Len(t, meta.Owners, 1)
	assert.Equal(t, "Owner Name", meta.Owners[0].DisplayName)
	assert.Equal(t, "owner@example.com", meta.Owners[0].EmailAddress)

	// Check last modifying user
	assert.NotNil(t, meta.LastModifyingUser)
	assert.Equal(t, "Last Modifier", meta.LastModifyingUser.DisplayName)
}

func TestParseFileToMetadata_NoOwners(t *testing.T) {
	driveFile := &drive.File{
		Id:   "file-no-owner",
		Name: "no-owner-file.txt",
	}

	meta := parseFileToMetadata(driveFile)

	assert.Equal(t, "file-no-owner", meta.ID)
	assert.Nil(t, meta.Owners)
	assert.Nil(t, meta.LastModifyingUser)
}

func TestParseFileToMetadata_NilLastModifier(t *testing.T) {
	driveFile := &drive.File{
		Id:                "file-no-modifier",
		Name:              "no-modifier-file.txt",
		LastModifyingUser: nil,
	}

	meta := parseFileToMetadata(driveFile)

	assert.Nil(t, meta.LastModifyingUser)
}
