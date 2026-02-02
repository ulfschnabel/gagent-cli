// Package drive provides Google Drive API operations.
package drive

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Service wraps the Drive API service.
type Service struct {
	svc *drive.Service
}

// NewService creates a new Drive service.
func NewService(ctx context.Context, client *http.Client) (*Service, error) {
	svc, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Drive service: %w", err)
	}
	return &Service{svc: svc}, nil
}

// FileSummary represents a summary of a Drive file.
type FileSummary struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	MimeType     string   `json:"mime_type"`
	Size         int64    `json:"size,omitempty"`
	ModifiedTime string   `json:"modified_time,omitempty"`
	Parents      []string `json:"parents,omitempty"`
	WebViewLink  string   `json:"web_view_link,omitempty"`
	Starred      bool     `json:"starred,omitempty"`
	Trashed      bool     `json:"trashed,omitempty"`
}

// FileMetadata represents full file metadata.
type FileMetadata struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	MimeType        string            `json:"mime_type"`
	Size            int64             `json:"size,omitempty"`
	CreatedTime     string            `json:"created_time,omitempty"`
	ModifiedTime    string            `json:"modified_time,omitempty"`
	Parents         []string          `json:"parents,omitempty"`
	WebViewLink     string            `json:"web_view_link,omitempty"`
	WebContentLink  string            `json:"web_content_link,omitempty"`
	IconLink        string            `json:"icon_link,omitempty"`
	Starred         bool              `json:"starred,omitempty"`
	Trashed         bool              `json:"trashed,omitempty"`
	Shared          bool              `json:"shared,omitempty"`
	Owners          []Owner           `json:"owners,omitempty"`
	LastModifyingUser *Owner          `json:"last_modifying_user,omitempty"`
	Description     string            `json:"description,omitempty"`
	Properties      map[string]string `json:"properties,omitempty"`
}

// Owner represents a file owner or modifier.
type Owner struct {
	DisplayName  string `json:"display_name,omitempty"`
	EmailAddress string `json:"email_address,omitempty"`
	PhotoLink    string `json:"photo_link,omitempty"`
}

// PermissionInfo represents file sharing permissions.
type PermissionInfo struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Role         string `json:"role"`
	EmailAddress string `json:"email_address,omitempty"`
	DisplayName  string `json:"display_name,omitempty"`
	Domain       string `json:"domain,omitempty"`
}

// QuotaInfo represents storage quota information.
type QuotaInfo struct {
	Limit             int64 `json:"limit,omitempty"`
	Usage             int64 `json:"usage"`
	UsageInDrive      int64 `json:"usage_in_drive"`
	UsageInTrash      int64 `json:"usage_in_trash"`
	UsageInDriveTrash int64 `json:"usage_in_drive_trash,omitempty"`
}

// CreateResult represents the result of creating a file or folder.
type CreateResult struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	WebViewLink string `json:"web_view_link,omitempty"`
}

// CopyResult represents the result of copying a file.
type CopyResult struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	WebViewLink string `json:"web_view_link,omitempty"`
}

// ShareResult represents the result of sharing a file.
type ShareResult struct {
	PermissionID string `json:"permission_id"`
	Role         string `json:"role"`
	Type         string `json:"type"`
}

// MimeTypes for common Drive file types.
const (
	MimeTypeFolder       = "application/vnd.google-apps.folder"
	MimeTypeDocument     = "application/vnd.google-apps.document"
	MimeTypeSpreadsheet  = "application/vnd.google-apps.spreadsheet"
	MimeTypePresentation = "application/vnd.google-apps.presentation"
)
