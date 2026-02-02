package drive

import (
	"fmt"

	"google.golang.org/api/drive/v3"
)

// CreateFolder creates a new folder.
func (s *Service) CreateFolder(name string, parentID string) (*CreateResult, error) {
	file := &drive.File{
		Name:     name,
		MimeType: MimeTypeFolder,
	}
	if parentID != "" {
		file.Parents = []string{parentID}
	}

	created, err := s.svc.Files.Create(file).
		Fields("id, name, webViewLink").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create folder: %w", err)
	}

	return &CreateResult{
		ID:          created.Id,
		Name:        created.Name,
		WebViewLink: created.WebViewLink,
	}, nil
}

// Delete permanently deletes a file or folder.
func (s *Service) Delete(fileID string) error {
	if err := s.svc.Files.Delete(fileID).Do(); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// Trash moves a file to trash.
func (s *Service) Trash(fileID string) error {
	_, err := s.svc.Files.Update(fileID, &drive.File{Trashed: true}).Do()
	if err != nil {
		return fmt.Errorf("failed to trash file: %w", err)
	}
	return nil
}

// Untrash restores a file from trash.
func (s *Service) Untrash(fileID string) error {
	_, err := s.svc.Files.Update(fileID, &drive.File{Trashed: false}).Do()
	if err != nil {
		return fmt.Errorf("failed to untrash file: %w", err)
	}
	return nil
}

// EmptyTrash permanently deletes all trashed files.
func (s *Service) EmptyTrash() error {
	if err := s.svc.Files.EmptyTrash().Do(); err != nil {
		return fmt.Errorf("failed to empty trash: %w", err)
	}
	return nil
}

// Move moves a file to a different folder.
func (s *Service) Move(fileID string, newParentID string) error {
	// Get current parents
	file, err := s.svc.Files.Get(fileID).Fields("parents").Do()
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}

	// Remove from old parents, add to new parent
	var oldParents string
	for i, p := range file.Parents {
		if i > 0 {
			oldParents += ","
		}
		oldParents += p
	}

	_, err = s.svc.Files.Update(fileID, nil).
		AddParents(newParentID).
		RemoveParents(oldParents).
		Do()
	if err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	return nil
}

// Copy copies a file.
func (s *Service) Copy(fileID string, newName string) (*CopyResult, error) {
	file := &drive.File{}
	if newName != "" {
		file.Name = newName
	}

	copied, err := s.svc.Files.Copy(fileID, file).
		Fields("id, name, webViewLink").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	return &CopyResult{
		ID:          copied.Id,
		Name:        copied.Name,
		WebViewLink: copied.WebViewLink,
	}, nil
}

// Rename renames a file.
func (s *Service) Rename(fileID string, newName string) error {
	_, err := s.svc.Files.Update(fileID, &drive.File{Name: newName}).Do()
	if err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}
	return nil
}

// ShareOptions contains options for sharing a file.
type ShareOptions struct {
	FileID       string
	EmailAddress string
	Role         string // reader, writer, commenter
	Type         string // user, group, domain, anyone
	SendEmail    bool
}

// Share shares a file with a user or group.
func (s *Service) Share(opts ShareOptions) (*ShareResult, error) {
	if opts.Role == "" {
		opts.Role = "reader"
	}
	if opts.Type == "" {
		opts.Type = "user"
	}

	permission := &drive.Permission{
		Type:         opts.Type,
		Role:         opts.Role,
		EmailAddress: opts.EmailAddress,
	}

	call := s.svc.Permissions.Create(opts.FileID, permission).
		Fields("id, role, type")

	if opts.SendEmail {
		call = call.SendNotificationEmail(true)
	}

	created, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to share file: %w", err)
	}

	return &ShareResult{
		PermissionID: created.Id,
		Role:         created.Role,
		Type:         created.Type,
	}, nil
}

// Unshare removes a permission from a file.
func (s *Service) Unshare(fileID string, permissionID string) error {
	if err := s.svc.Permissions.Delete(fileID, permissionID).Do(); err != nil {
		return fmt.Errorf("failed to unshare file: %w", err)
	}
	return nil
}

// Star stars a file.
func (s *Service) Star(fileID string) error {
	_, err := s.svc.Files.Update(fileID, &drive.File{Starred: true}).Do()
	if err != nil {
		return fmt.Errorf("failed to star file: %w", err)
	}
	return nil
}

// Unstar removes star from a file.
func (s *Service) Unstar(fileID string) error {
	_, err := s.svc.Files.Update(fileID, &drive.File{Starred: false}).Do()
	if err != nil {
		return fmt.Errorf("failed to unstar file: %w", err)
	}
	return nil
}

// UpdateDescription updates a file's description.
func (s *Service) UpdateDescription(fileID string, description string) error {
	_, err := s.svc.Files.Update(fileID, &drive.File{Description: description}).Do()
	if err != nil {
		return fmt.Errorf("failed to update description: %w", err)
	}
	return nil
}
