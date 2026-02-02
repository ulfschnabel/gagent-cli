package main

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/ulfhaga/gagent-cli/internal/auth"
	"github.com/ulfhaga/gagent-cli/internal/config"
	"github.com/ulfhaga/gagent-cli/internal/drive"
	"github.com/ulfhaga/gagent-cli/internal/output"
)

// driveReadService creates a Drive service with read scope.
func driveReadService(ctx context.Context) (*drive.Service, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, err
	}

	if err := auth.RequireScope(configDir, auth.ScopeRead); err != nil {
		return nil, err
	}

	client, err := auth.GetClient(ctx, configDir, cfg.ClientID, cfg.ClientSecret, auth.ScopeRead)
	if err != nil {
		return nil, err
	}

	return drive.NewService(ctx, client)
}

// driveWriteService creates a Drive service with write scope.
func driveWriteService(ctx context.Context) (*drive.Service, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, err
	}

	if err := auth.RequireScope(configDir, auth.ScopeWrite); err != nil {
		return nil, err
	}

	client, err := auth.GetClient(ctx, configDir, cfg.ClientID, cfg.ClientSecret, auth.ScopeWrite)
	if err != nil {
		return nil, err
	}

	return drive.NewService(ctx, client)
}

func driveListCmd() *cobra.Command {
	var limit int64
	var folder string
	var mimeType string
	var query string
	var pageToken string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List files in Drive",
		Long:  "Returns files in Drive with optional filtering.",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := driveReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			files, nextToken, err := svc.List(drive.ListOptions{
				Query:      query,
				MimeType:   mimeType,
				ParentID:   folder,
				MaxResults: limit,
				PageToken:  pageToken,
			})
			if err != nil {
				output.APIError(err)
				return
			}

			result := map[string]interface{}{
				"files": files,
				"count": len(files),
			}
			if nextToken != "" {
				result["next_page_token"] = nextToken
			}

			output.Success(result, "read")
		},
	}

	cmd.Flags().Int64Var(&limit, "limit", 20, "Maximum number of files to return")
	cmd.Flags().StringVar(&folder, "folder", "", "Parent folder ID to list")
	cmd.Flags().StringVar(&mimeType, "mime-type", "", "Filter by MIME type")
	cmd.Flags().StringVar(&query, "query", "", "Custom query string")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Page token for pagination")

	return cmd
}

func driveGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <file-id>",
		Short: "Get file metadata",
		Long:  "Returns full metadata for a specific file.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := driveReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			file, err := svc.Get(args[0])
			if err != nil {
				output.NotFoundError("File", args[0])
				return
			}

			output.Success(file, "read")
		},
	}

	return cmd
}

func driveSearchCmd() *cobra.Command {
	var limit int64

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search files by content",
		Long:  "Full-text search across file names and content.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := driveReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			files, err := svc.Search(args[0], limit)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"files": files,
				"count": len(files),
				"query": args[0],
			}, "read")
		},
	}

	cmd.Flags().Int64Var(&limit, "limit", 20, "Maximum number of results")

	return cmd
}

func driveFoldersCmd() *cobra.Command {
	var parent string
	var limit int64

	cmd := &cobra.Command{
		Use:   "folders",
		Short: "List folders",
		Long:  "List folders in Drive.",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := driveReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			folders, err := svc.ListFolders(parent, limit)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"folders": folders,
				"count":   len(folders),
			}, "read")
		},
	}

	cmd.Flags().StringVar(&parent, "parent", "", "Parent folder ID")
	cmd.Flags().Int64Var(&limit, "limit", 50, "Maximum number of folders")

	return cmd
}

func drivePermissionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "permissions <file-id>",
		Short: "List file permissions",
		Long:  "List sharing permissions for a file.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := driveReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			permissions, err := svc.GetPermissions(args[0])
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"permissions": permissions,
				"file_id":     args[0],
			}, "read")
		},
	}

	return cmd
}

func driveQuotaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quota",
		Short: "Show storage quota",
		Long:  "Display storage quota information.",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := driveReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			quota, err := svc.GetQuota()
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(quota, "read")
		},
	}

	return cmd
}

func driveTrashListCmd() *cobra.Command {
	var limit int64

	cmd := &cobra.Command{
		Use:   "trash-list",
		Short: "List trashed files",
		Long:  "List files in trash.",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := driveReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			files, err := svc.ListTrashed(limit)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"files": files,
				"count": len(files),
			}, "read")
		},
	}

	cmd.Flags().Int64Var(&limit, "limit", 20, "Maximum number of files")

	return cmd
}

// Write commands

func driveCreateFolderCmd() *cobra.Command {
	var parent string

	cmd := &cobra.Command{
		Use:   "create-folder <name>",
		Short: "Create a folder",
		Long:  "Create a new folder in Drive.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := driveWriteService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrScopeInsufficient, err)
				return
			}

			result, err := svc.CreateFolder(args[0], parent)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&parent, "parent", "", "Parent folder ID")

	return cmd
}

func driveDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <file-id>",
		Short: "Permanently delete a file",
		Long:  "Permanently delete a file or folder (cannot be undone).",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := driveWriteService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrScopeInsufficient, err)
				return
			}

			if err := svc.Delete(args[0]); err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"deleted": true,
				"file_id": args[0],
			}, "write")
		},
	}

	return cmd
}

func driveTrashCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trash <file-id>",
		Short: "Move file to trash",
		Long:  "Move a file to trash (can be restored).",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := driveWriteService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrScopeInsufficient, err)
				return
			}

			if err := svc.Trash(args[0]); err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"trashed": true,
				"file_id": args[0],
			}, "write")
		},
	}

	return cmd
}

func driveUntrashCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "untrash <file-id>",
		Short: "Restore file from trash",
		Long:  "Restore a file from trash.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := driveWriteService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrScopeInsufficient, err)
				return
			}

			if err := svc.Untrash(args[0]); err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"restored": true,
				"file_id":  args[0],
			}, "write")
		},
	}

	return cmd
}

func driveEmptyTrashCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "empty-trash",
		Short: "Empty trash",
		Long:  "Permanently delete all files in trash.",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := driveWriteService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrScopeInsufficient, err)
				return
			}

			if err := svc.EmptyTrash(); err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"emptied": true,
			}, "write")
		},
	}

	return cmd
}

func driveMoveCmd() *cobra.Command {
	var toFolder string

	cmd := &cobra.Command{
		Use:   "move <file-id>",
		Short: "Move file to folder",
		Long:  "Move a file to a different folder.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := driveWriteService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrScopeInsufficient, err)
				return
			}

			if err := svc.Move(args[0], toFolder); err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"moved":     true,
				"file_id":   args[0],
				"to_folder": toFolder,
			}, "write")
		},
	}

	cmd.Flags().StringVar(&toFolder, "to", "", "Destination folder ID")
	cmd.MarkFlagRequired("to")

	return cmd
}

func driveCopyCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "copy <file-id>",
		Short: "Copy a file",
		Long:  "Create a copy of a file.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := driveWriteService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrScopeInsufficient, err)
				return
			}

			result, err := svc.Copy(args[0], name)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Name for the copy")

	return cmd
}

func driveRenameCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "rename <file-id>",
		Short: "Rename a file",
		Long:  "Rename a file or folder.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := driveWriteService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrScopeInsufficient, err)
				return
			}

			if err := svc.Rename(args[0], name); err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"renamed":  true,
				"file_id":  args[0],
				"new_name": name,
			}, "write")
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "New name")
	cmd.MarkFlagRequired("name")

	return cmd
}

func driveShareCmd() *cobra.Command {
	var email string
	var role string
	var sendEmail bool

	cmd := &cobra.Command{
		Use:   "share <file-id>",
		Short: "Share a file",
		Long:  "Share a file with a user.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := driveWriteService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrScopeInsufficient, err)
				return
			}

			result, err := svc.Share(drive.ShareOptions{
				FileID:       args[0],
				EmailAddress: email,
				Role:         role,
				Type:         "user",
				SendEmail:    sendEmail,
			})
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "Email address to share with")
	cmd.Flags().StringVar(&role, "role", "reader", "Role: reader, writer, or commenter")
	cmd.Flags().BoolVar(&sendEmail, "notify", true, "Send notification email")
	cmd.MarkFlagRequired("email")

	return cmd
}

func driveUnshareCmd() *cobra.Command {
	var permissionID string

	cmd := &cobra.Command{
		Use:   "unshare <file-id>",
		Short: "Remove sharing permission",
		Long:  "Remove a sharing permission from a file.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := driveWriteService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrScopeInsufficient, err)
				return
			}

			if err := svc.Unshare(args[0], permissionID); err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"unshared":      true,
				"file_id":       args[0],
				"permission_id": permissionID,
			}, "write")
		},
	}

	cmd.Flags().StringVar(&permissionID, "permission-id", "", "Permission ID to remove")
	cmd.MarkFlagRequired("permission-id")

	return cmd
}

func driveStarCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "star <file-id>",
		Short: "Star a file",
		Long:  "Add a file to starred.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := driveWriteService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrScopeInsufficient, err)
				return
			}

			if err := svc.Star(args[0]); err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"starred": true,
				"file_id": args[0],
			}, "write")
		},
	}

	return cmd
}

func driveUnstarCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unstar <file-id>",
		Short: "Unstar a file",
		Long:  "Remove a file from starred.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := driveWriteService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrScopeInsufficient, err)
				return
			}

			if err := svc.Unstar(args[0]); err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"unstarred": true,
				"file_id":   args[0],
			}, "write")
		},
	}

	return cmd
}

// API commands for low-level access

func driveAPICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Low-level Drive API commands",
		Long:  "Direct access to Drive API operations.",
	}

	cmd.AddCommand(driveAPIListCmd())
	cmd.AddCommand(driveAPIGetCmd())

	return cmd
}

func driveAPIListCmd() *cobra.Command {
	var pageSize int64
	var pageToken string
	var query string
	var orderBy string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List files with raw options",
		Long:  "List files with full API control.",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := driveReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			files, nextToken, err := svc.List(drive.ListOptions{
				Query:      query,
				MaxResults: pageSize,
				PageToken:  pageToken,
				OrderBy:    orderBy,
				Trashed:    true, // Include trashed in API mode
			})
			if err != nil {
				output.APIError(err)
				return
			}

			result := map[string]interface{}{
				"files": files,
				"count": len(files),
			}
			if nextToken != "" {
				result["next_page_token"] = nextToken
			}

			output.Success(result, "read")
		},
	}

	cmd.Flags().Int64Var(&pageSize, "page-size", 100, "Number of files per page")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Page token")
	cmd.Flags().StringVar(&query, "q", "", "Query string (Drive API format)")
	cmd.Flags().StringVar(&orderBy, "order-by", "", "Order by field")

	return cmd
}

func driveAPIGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <file-id>",
		Short: "Get file metadata",
		Long:  "Get full file metadata.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := driveReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			file, err := svc.Get(args[0])
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(file, "read")
		},
	}

	return cmd
}

// driveCmd returns the drive command group - called from main.go
func driveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drive",
		Short: "Google Drive commands",
		Long:  "Manage files and folders in Google Drive.",
	}

	// Read commands
	cmd.AddCommand(driveListCmd())
	cmd.AddCommand(driveGetCmd())
	cmd.AddCommand(driveSearchCmd())
	cmd.AddCommand(driveFoldersCmd())
	cmd.AddCommand(drivePermissionsCmd())
	cmd.AddCommand(driveQuotaCmd())
	cmd.AddCommand(driveTrashListCmd())

	// Write commands
	cmd.AddCommand(driveCreateFolderCmd())
	cmd.AddCommand(driveDeleteCmd())
	cmd.AddCommand(driveTrashCmd())
	cmd.AddCommand(driveUntrashCmd())
	cmd.AddCommand(driveEmptyTrashCmd())
	cmd.AddCommand(driveMoveCmd())
	cmd.AddCommand(driveCopyCmd())
	cmd.AddCommand(driveRenameCmd())
	cmd.AddCommand(driveShareCmd())
	cmd.AddCommand(driveUnshareCmd())
	cmd.AddCommand(driveStarCmd())
	cmd.AddCommand(driveUnstarCmd())

	// API commands
	cmd.AddCommand(driveAPICmd())

	return cmd
}
