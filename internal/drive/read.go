package drive

import (
	"fmt"

	"google.golang.org/api/drive/v3"
)

// ListOptions contains options for listing files.
type ListOptions struct {
	Query      string
	MimeType   string
	ParentID   string
	MaxResults int64
	PageToken  string
	Trashed    bool
	OrderBy    string
}

// List returns files matching the criteria.
func (s *Service) List(opts ListOptions) ([]FileSummary, string, error) {
	call := s.svc.Files.List().
		Fields("nextPageToken, files(id, name, mimeType, size, modifiedTime, parents, webViewLink, starred, trashed)")

	// Build query
	var queryParts []string

	if opts.Query != "" {
		queryParts = append(queryParts, opts.Query)
	}
	if opts.MimeType != "" {
		queryParts = append(queryParts, fmt.Sprintf("mimeType = '%s'", opts.MimeType))
	}
	if opts.ParentID != "" {
		queryParts = append(queryParts, fmt.Sprintf("'%s' in parents", opts.ParentID))
	}
	if !opts.Trashed {
		queryParts = append(queryParts, "trashed = false")
	}

	if len(queryParts) > 0 {
		q := queryParts[0]
		for i := 1; i < len(queryParts); i++ {
			q += " and " + queryParts[i]
		}
		call = call.Q(q)
	}

	if opts.MaxResults > 0 {
		call = call.PageSize(opts.MaxResults)
	} else {
		call = call.PageSize(20)
	}
	if opts.PageToken != "" {
		call = call.PageToken(opts.PageToken)
	}
	if opts.OrderBy != "" {
		call = call.OrderBy(opts.OrderBy)
	} else {
		call = call.OrderBy("modifiedTime desc")
	}

	resp, err := call.Do()
	if err != nil {
		return nil, "", fmt.Errorf("failed to list files: %w", err)
	}

	files := make([]FileSummary, 0, len(resp.Files))
	for _, f := range resp.Files {
		files = append(files, parseFileToSummary(f))
	}

	return files, resp.NextPageToken, nil
}

// Get returns metadata for a specific file.
func (s *Service) Get(fileID string) (*FileMetadata, error) {
	file, err := s.svc.Files.Get(fileID).
		Fields("id, name, mimeType, size, createdTime, modifiedTime, parents, webViewLink, webContentLink, iconLink, starred, trashed, shared, owners, lastModifyingUser, description, properties").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return parseFileToMetadata(file), nil
}

// Search performs a full-text search.
func (s *Service) Search(query string, limit int64) ([]FileSummary, error) {
	if limit <= 0 {
		limit = 20
	}

	files, _, err := s.List(ListOptions{
		Query:      fmt.Sprintf("fullText contains '%s'", query),
		MaxResults: limit,
	})
	return files, err
}

// ListFolders returns all folders.
func (s *Service) ListFolders(parentID string, limit int64) ([]FileSummary, error) {
	if limit <= 0 {
		limit = 50
	}

	files, _, err := s.List(ListOptions{
		MimeType:   MimeTypeFolder,
		ParentID:   parentID,
		MaxResults: limit,
		OrderBy:    "name",
	})
	return files, err
}

// GetPermissions returns sharing permissions for a file.
func (s *Service) GetPermissions(fileID string) ([]PermissionInfo, error) {
	resp, err := s.svc.Permissions.List(fileID).
		Fields("permissions(id, type, role, emailAddress, displayName, domain)").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	permissions := make([]PermissionInfo, 0, len(resp.Permissions))
	for _, p := range resp.Permissions {
		permissions = append(permissions, PermissionInfo{
			ID:           p.Id,
			Type:         p.Type,
			Role:         p.Role,
			EmailAddress: p.EmailAddress,
			DisplayName:  p.DisplayName,
			Domain:       p.Domain,
		})
	}

	return permissions, nil
}

// GetQuota returns storage quota information.
func (s *Service) GetQuota() (*QuotaInfo, error) {
	about, err := s.svc.About.Get().
		Fields("storageQuota").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get quota: %w", err)
	}

	return &QuotaInfo{
		Limit:             about.StorageQuota.Limit,
		Usage:             about.StorageQuota.Usage,
		UsageInDrive:      about.StorageQuota.UsageInDrive,
		UsageInTrash:      about.StorageQuota.UsageInDriveTrash,
	}, nil
}

// ListTrashed returns files in trash.
func (s *Service) ListTrashed(limit int64) ([]FileSummary, error) {
	if limit <= 0 {
		limit = 20
	}

	call := s.svc.Files.List().
		Q("trashed = true").
		Fields("files(id, name, mimeType, size, modifiedTime, parents, webViewLink, starred, trashed)").
		PageSize(limit).
		OrderBy("modifiedTime desc")

	resp, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list trashed files: %w", err)
	}

	files := make([]FileSummary, 0, len(resp.Files))
	for _, f := range resp.Files {
		files = append(files, parseFileToSummary(f))
	}

	return files, nil
}

// parseFileToSummary converts a Drive file to a FileSummary.
func parseFileToSummary(f *drive.File) FileSummary {
	return FileSummary{
		ID:           f.Id,
		Name:         f.Name,
		MimeType:     f.MimeType,
		Size:         f.Size,
		ModifiedTime: f.ModifiedTime,
		Parents:      f.Parents,
		WebViewLink:  f.WebViewLink,
		Starred:      f.Starred,
		Trashed:      f.Trashed,
	}
}

// parseFileToMetadata converts a Drive file to FileMetadata.
func parseFileToMetadata(f *drive.File) *FileMetadata {
	meta := &FileMetadata{
		ID:             f.Id,
		Name:           f.Name,
		MimeType:       f.MimeType,
		Size:           f.Size,
		CreatedTime:    f.CreatedTime,
		ModifiedTime:   f.ModifiedTime,
		Parents:        f.Parents,
		WebViewLink:    f.WebViewLink,
		WebContentLink: f.WebContentLink,
		IconLink:       f.IconLink,
		Starred:        f.Starred,
		Trashed:        f.Trashed,
		Shared:         f.Shared,
		Description:    f.Description,
		Properties:     f.Properties,
	}

	for _, o := range f.Owners {
		meta.Owners = append(meta.Owners, Owner{
			DisplayName:  o.DisplayName,
			EmailAddress: o.EmailAddress,
			PhotoLink:    o.PhotoLink,
		})
	}

	if f.LastModifyingUser != nil {
		meta.LastModifyingUser = &Owner{
			DisplayName:  f.LastModifyingUser.DisplayName,
			EmailAddress: f.LastModifyingUser.EmailAddress,
			PhotoLink:    f.LastModifyingUser.PhotoLink,
		}
	}

	return meta
}
