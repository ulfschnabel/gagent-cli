package drive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMimeTypeConstants(t *testing.T) {
	assert.Equal(t, "application/vnd.google-apps.folder", MimeTypeFolder)
	assert.Equal(t, "application/vnd.google-apps.document", MimeTypeDocument)
	assert.Equal(t, "application/vnd.google-apps.spreadsheet", MimeTypeSpreadsheet)
	assert.Equal(t, "application/vnd.google-apps.presentation", MimeTypePresentation)
}

func TestFileSummaryJSON(t *testing.T) {
	file := FileSummary{
		ID:           "test-id",
		Name:         "test-file.txt",
		MimeType:     "text/plain",
		Size:         1024,
		ModifiedTime: "2024-01-15T10:30:00Z",
		Parents:      []string{"parent-folder-id"},
		WebViewLink:  "https://drive.google.com/file/d/test-id",
		Starred:      true,
		Trashed:      false,
	}

	assert.Equal(t, "test-id", file.ID)
	assert.Equal(t, "test-file.txt", file.Name)
	assert.Equal(t, "text/plain", file.MimeType)
	assert.Equal(t, int64(1024), file.Size)
	assert.True(t, file.Starred)
	assert.False(t, file.Trashed)
	assert.Len(t, file.Parents, 1)
}

func TestFileMetadataJSON(t *testing.T) {
	meta := FileMetadata{
		ID:             "test-id",
		Name:           "test-file.txt",
		MimeType:       "text/plain",
		Size:           2048,
		CreatedTime:    "2024-01-10T08:00:00Z",
		ModifiedTime:   "2024-01-15T10:30:00Z",
		WebViewLink:    "https://drive.google.com/file/d/test-id",
		WebContentLink: "https://drive.google.com/uc?id=test-id",
		Shared:         true,
		Description:    "Test file description",
		Owners: []Owner{
			{DisplayName: "Test User", EmailAddress: "test@example.com"},
		},
	}

	assert.Equal(t, "test-id", meta.ID)
	assert.Equal(t, "test-file.txt", meta.Name)
	assert.True(t, meta.Shared)
	assert.Equal(t, "Test file description", meta.Description)
	assert.Len(t, meta.Owners, 1)
	assert.Equal(t, "Test User", meta.Owners[0].DisplayName)
}

func TestPermissionInfo(t *testing.T) {
	perm := PermissionInfo{
		ID:           "perm-123",
		Type:         "user",
		Role:         "writer",
		EmailAddress: "collaborator@example.com",
		DisplayName:  "Collaborator",
	}

	assert.Equal(t, "perm-123", perm.ID)
	assert.Equal(t, "user", perm.Type)
	assert.Equal(t, "writer", perm.Role)
	assert.Equal(t, "collaborator@example.com", perm.EmailAddress)
}

func TestQuotaInfo(t *testing.T) {
	quota := QuotaInfo{
		Limit:        15000000000, // 15 GB
		Usage:        5000000000,  // 5 GB
		UsageInDrive: 3000000000,  // 3 GB
		UsageInTrash: 500000000,   // 500 MB
	}

	assert.Equal(t, int64(15000000000), quota.Limit)
	assert.Equal(t, int64(5000000000), quota.Usage)
	assert.Equal(t, int64(3000000000), quota.UsageInDrive)
	assert.Equal(t, int64(500000000), quota.UsageInTrash)
}

func TestCreateResult(t *testing.T) {
	result := CreateResult{
		ID:          "new-folder-id",
		Name:        "New Folder",
		WebViewLink: "https://drive.google.com/drive/folders/new-folder-id",
	}

	assert.Equal(t, "new-folder-id", result.ID)
	assert.Equal(t, "New Folder", result.Name)
	assert.Contains(t, result.WebViewLink, "new-folder-id")
}

func TestCopyResult(t *testing.T) {
	result := CopyResult{
		ID:          "copied-file-id",
		Name:        "Copy of Document",
		WebViewLink: "https://drive.google.com/file/d/copied-file-id",
	}

	assert.Equal(t, "copied-file-id", result.ID)
	assert.Equal(t, "Copy of Document", result.Name)
}

func TestShareResult(t *testing.T) {
	result := ShareResult{
		PermissionID: "perm-new-123",
		Role:         "reader",
		Type:         "user",
	}

	assert.Equal(t, "perm-new-123", result.PermissionID)
	assert.Equal(t, "reader", result.Role)
	assert.Equal(t, "user", result.Type)
}

func TestOwner(t *testing.T) {
	owner := Owner{
		DisplayName:  "John Doe",
		EmailAddress: "john@example.com",
		PhotoLink:    "https://example.com/photo.jpg",
	}

	assert.Equal(t, "John Doe", owner.DisplayName)
	assert.Equal(t, "john@example.com", owner.EmailAddress)
	assert.Equal(t, "https://example.com/photo.jpg", owner.PhotoLink)
}
