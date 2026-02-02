package main

import (
	"context"
	"encoding/base64"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ulfhaga/gagent-cli/internal/auth"
	"github.com/ulfhaga/gagent-cli/internal/config"
	"github.com/ulfhaga/gagent-cli/internal/gmail"
	"github.com/ulfhaga/gagent-cli/internal/output"
)

// gmailReadService creates a Gmail service with read scope.
func gmailReadService(ctx context.Context) (*gmail.Service, error) {
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

	return gmail.NewService(ctx, client)
}

// gmailWriteService creates a Gmail service with write scope.
func gmailWriteService(ctx context.Context) (*gmail.Service, error) {
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

	return gmail.NewService(ctx, client)
}

func gmailInboxCmd() *cobra.Command {
	var limit int64
	var unreadOnly bool

	cmd := &cobra.Command{
		Use:   "inbox",
		Short: "List inbox messages",
		Long:  "Returns recent inbox messages with subject, from, date, snippet.",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := gmailReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			messages, err := svc.Inbox(limit, unreadOnly)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"messages": messages,
				"count":    len(messages),
			}, "read")
		},
	}

	cmd.Flags().Int64VarP(&limit, "limit", "n", 10, "Maximum number of messages to return")
	cmd.Flags().BoolVar(&unreadOnly, "unread-only", false, "Only return unread messages")

	return cmd
}

func gmailReadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "read <message-id>",
		Short: "Read a message",
		Long:  "Returns full message: headers, body (plain + html), attachments list.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := gmailReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			msg, err := svc.Get(args[0])
			if err != nil {
				output.NotFoundError("Message", args[0])
				return
			}

			output.Success(msg, "read")
		},
	}
}

func gmailSearchCmd() *cobra.Command {
	var limit int64

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search messages",
		Long:  "Search using Gmail query syntax, returns matching messages.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := gmailReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			messages, err := svc.Search(args[0], limit)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"query":    args[0],
				"messages": messages,
				"count":    len(messages),
			}, "read")
		},
	}

	cmd.Flags().Int64VarP(&limit, "limit", "n", 10, "Maximum number of messages to return")

	return cmd
}

func gmailThreadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "thread <thread-id>",
		Short: "Get a conversation thread",
		Long:  "Returns all messages in a conversation thread.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := gmailReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			thread, err := svc.GetThread(args[0])
			if err != nil {
				output.NotFoundError("Thread", args[0])
				return
			}

			output.Success(thread, "read")
		},
	}
}

func gmailSendCmd() *cobra.Command {
	var to, cc, bcc []string
	var subject, body string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send an email",
		Long: `Compose and send a NEW email in one step.

NOTE: This creates a standalone email that will NOT appear in an existing
conversation thread. To reply to an email and keep it in the same thread,
use 'gmail reply' instead, which sets the proper In-Reply-To headers.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(to) == 0 {
				output.InvalidInputError("At least one recipient is required (--to)")
				return
			}

			opts := gmail.SendOptions{
				To:      to,
				Cc:      cc,
				Bcc:     bcc,
				Subject: subject,
				Body:    body,
			}

			if dryRun {
				output.SuccessNoScope(map[string]interface{}{
					"dry_run": true,
					"to":      to,
					"cc":      cc,
					"bcc":     bcc,
					"subject": subject,
					"body":    body,
				})
				return
			}

			ctx := context.Background()
			svc, err := gmailWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.Send(opts)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringSliceVar(&to, "to", nil, "Recipient email addresses (required)")
	cmd.Flags().StringSliceVar(&cc, "cc", nil, "CC recipients")
	cmd.Flags().StringSliceVar(&bcc, "bcc", nil, "BCC recipients")
	cmd.Flags().StringVar(&subject, "subject", "", "Email subject (required)")
	cmd.Flags().StringVar(&body, "body", "", "Email body (required)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be sent without actually sending")

	cmd.MarkFlagRequired("to")
	cmd.MarkFlagRequired("subject")
	cmd.MarkFlagRequired("body")

	return cmd
}

func gmailReplyCmd() *cobra.Command {
	var body string
	var replyAll bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "reply <message-id>",
		Short: "Reply to a message",
		Long:  "Fetches original, sets In-Reply-To/References headers, sends reply.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			if dryRun {
				// For dry-run, we need to read the original message first
				svc, err := gmailReadService(ctx)
				if err != nil {
					output.FailureFromError(output.ErrAuthRequired, err)
					return
				}

				original, err := svc.Get(args[0])
				if err != nil {
					output.NotFoundError("Message", args[0])
					return
				}

				output.SuccessNoScope(map[string]interface{}{
					"dry_run":          true,
					"original_from":    original.From,
					"original_subject": original.Subject,
					"reply_all":        replyAll,
					"body":             body,
				})
				return
			}

			svc, err := gmailWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.Reply(gmail.ReplyOptions{
				MessageID: args[0],
				Body:      body,
				ReplyAll:  replyAll,
			})
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&body, "body", "", "Reply body (required)")
	cmd.Flags().BoolVar(&replyAll, "reply-all", false, "Reply to all recipients")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be sent without actually sending")

	cmd.MarkFlagRequired("body")

	return cmd
}

func gmailForwardCmd() *cobra.Command {
	var to []string
	var body string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "forward <message-id>",
		Short: "Forward a message",
		Long:  "Fetches original, includes quoted content, sends to new recipient.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(to) == 0 {
				output.InvalidInputError("At least one recipient is required (--to)")
				return
			}

			ctx := context.Background()

			if dryRun {
				svc, err := gmailReadService(ctx)
				if err != nil {
					output.FailureFromError(output.ErrAuthRequired, err)
					return
				}

				original, err := svc.Get(args[0])
				if err != nil {
					output.NotFoundError("Message", args[0])
					return
				}

				output.SuccessNoScope(map[string]interface{}{
					"dry_run":          true,
					"to":               to,
					"original_from":    original.From,
					"original_subject": original.Subject,
					"body":             body,
				})
				return
			}

			svc, err := gmailWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.Forward(gmail.ForwardOptions{
				MessageID: args[0],
				To:        to,
				Body:      body,
			})
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringSliceVar(&to, "to", nil, "Forward recipients (required)")
	cmd.Flags().StringVar(&body, "body", "", "Optional additional message")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be sent without actually sending")

	cmd.MarkFlagRequired("to")

	return cmd
}

func gmailDraftCmd() *cobra.Command {
	var to, cc, bcc []string
	var subject, body string

	cmd := &cobra.Command{
		Use:   "draft",
		Short: "Create a draft",
		Long: `Creates a NEW email draft, returns draft ID for later editing/sending.

NOTE: This creates a standalone email that will NOT appear in an existing
conversation thread. To reply to an email and keep it in the same thread,
use 'gmail reply' instead, which sets the proper In-Reply-To headers.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(to) == 0 {
				output.InvalidInputError("At least one recipient is required (--to)")
				return
			}

			ctx := context.Background()
			svc, err := gmailWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.DraftCreate(gmail.SendOptions{
				To:      to,
				Cc:      cc,
				Bcc:     bcc,
				Subject: subject,
				Body:    body,
			})
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringSliceVar(&to, "to", nil, "Recipient email addresses (required)")
	cmd.Flags().StringSliceVar(&cc, "cc", nil, "CC recipients")
	cmd.Flags().StringSliceVar(&bcc, "bcc", nil, "BCC recipients")
	cmd.Flags().StringVar(&subject, "subject", "", "Email subject (required)")
	cmd.Flags().StringVar(&body, "body", "", "Email body (required)")

	cmd.MarkFlagRequired("to")
	cmd.MarkFlagRequired("subject")
	cmd.MarkFlagRequired("body")

	return cmd
}

// Gmail API commands
func gmailAPICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Low-level Gmail API commands",
		Long:  "Direct access to Gmail API operations.",
	}

	cmd.AddCommand(gmailAPIListCmd())
	cmd.AddCommand(gmailAPIGetCmd())
	cmd.AddCommand(gmailAPILabelsCmd())
	cmd.AddCommand(gmailAPIAttachmentCmd())
	cmd.AddCommand(gmailAPIModifyCmd())
	cmd.AddCommand(gmailAPITrashCmd())
	cmd.AddCommand(gmailAPIUntrashCmd())
	cmd.AddCommand(gmailAPIDeleteCmd())
	cmd.AddCommand(gmailAPISendRawCmd())
	cmd.AddCommand(gmailAPIDraftCreateCmd())
	cmd.AddCommand(gmailAPIDraftSendCmd())
	cmd.AddCommand(gmailAPIDraftUpdateCmd())
	cmd.AddCommand(gmailAPIDraftDeleteCmd())

	return cmd
}

func gmailAPIListCmd() *cobra.Command {
	var label, query, pageToken string
	var limit int64

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List messages",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := gmailReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			opts := gmail.ListOptions{
				Query:      query,
				MaxResults: limit,
				PageToken:  pageToken,
			}
			if label != "" {
				opts.LabelIDs = []string{label}
			}

			messages, nextPageToken, err := svc.List(opts)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"messages":        messages,
				"count":           len(messages),
				"next_page_token": nextPageToken,
			}, "read")
		},
	}

	cmd.Flags().StringVar(&label, "label", "", "Label ID to filter by")
	cmd.Flags().StringVar(&query, "query", "", "Gmail search query")
	cmd.Flags().Int64VarP(&limit, "limit", "n", 10, "Maximum number of messages")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Page token for pagination")

	return cmd
}

func gmailAPIGetCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "get <message-id>",
		Short: "Get a message",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := gmailReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			if format == "raw" {
				raw, err := svc.GetRaw(args[0])
				if err != nil {
					output.NotFoundError("Message", args[0])
					return
				}
				output.Success(map[string]interface{}{
					"id":  args[0],
					"raw": base64.StdEncoding.EncodeToString(raw),
				}, "read")
				return
			}

			msg, err := svc.Get(args[0])
			if err != nil {
				output.NotFoundError("Message", args[0])
				return
			}

			output.Success(msg, "read")
		},
	}

	cmd.Flags().StringVar(&format, "format", "full", "Format: full, metadata, minimal, raw")

	return cmd
}

func gmailAPILabelsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "labels",
		Short: "List labels",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := gmailReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			labels, err := svc.Labels()
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"labels": labels,
				"count":  len(labels),
			}, "read")
		},
	}
}

func gmailAPIAttachmentCmd() *cobra.Command {
	var saveTo string

	cmd := &cobra.Command{
		Use:   "attachment <message-id> <attachment-id>",
		Short: "Get an attachment",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := gmailReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			data, err := svc.GetAttachment(args[0], args[1])
			if err != nil {
				output.NotFoundError("Attachment", args[1])
				return
			}

			if saveTo != "" {
				if err := os.WriteFile(saveTo, data, 0644); err != nil {
					output.FailureFromError(output.ErrInternal, err)
					return
				}
				output.Success(map[string]interface{}{
					"saved_to": saveTo,
					"size":     len(data),
				}, "read")
				return
			}

			output.Success(map[string]interface{}{
				"data": base64.StdEncoding.EncodeToString(data),
				"size": len(data),
			}, "read")
		},
	}

	cmd.Flags().StringVar(&saveTo, "save-to", "", "File path to save attachment")

	return cmd
}

func gmailAPIModifyCmd() *cobra.Command {
	var addLabels, removeLabels string

	cmd := &cobra.Command{
		Use:   "modify <message-id>",
		Short: "Modify message labels",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := gmailWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			var add, remove []string
			if addLabels != "" {
				add = strings.Split(addLabels, ",")
			}
			if removeLabels != "" {
				remove = strings.Split(removeLabels, ",")
			}

			if err := svc.ModifyLabels(args[0], add, remove); err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"message_id":     args[0],
				"added_labels":   add,
				"removed_labels": remove,
			}, "write")
		},
	}

	cmd.Flags().StringVar(&addLabels, "add-labels", "", "Labels to add (comma-separated)")
	cmd.Flags().StringVar(&removeLabels, "remove-labels", "", "Labels to remove (comma-separated)")

	return cmd
}

func gmailAPITrashCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "trash <message-id>",
		Short: "Move message to trash",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := gmailWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			if err := svc.Trash(args[0]); err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"message_id": args[0],
				"trashed":    true,
			}, "write")
		},
	}
}

func gmailAPIUntrashCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "untrash <message-id>",
		Short: "Remove message from trash",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := gmailWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			if err := svc.Untrash(args[0]); err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"message_id": args[0],
				"untrashed":  true,
			}, "write")
		},
	}
}

func gmailAPIDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <message-id>",
		Short: "Permanently delete a message",
		Long:  "Permanently deletes a message. This action cannot be undone.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := gmailWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			if err := svc.Delete(args[0]); err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"message_id": args[0],
				"deleted":    true,
			}, "write")
		},
	}
}

func gmailAPISendRawCmd() *cobra.Command {
	var raw string

	cmd := &cobra.Command{
		Use:   "send-raw",
		Short: "Send a raw RFC 2822 message",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := gmailWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			decoded, err := base64.StdEncoding.DecodeString(raw)
			if err != nil {
				output.InvalidInputError("Invalid base64 encoding for raw message")
				return
			}

			result, err := svc.SendRaw(decoded)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&raw, "raw", "", "Base64-encoded RFC 2822 message (required)")
	cmd.MarkFlagRequired("raw")

	return cmd
}

func gmailAPIDraftCreateCmd() *cobra.Command {
	var raw string

	cmd := &cobra.Command{
		Use:   "draft-create",
		Short: "Create a draft from raw message",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := gmailWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			// For simplicity, we'll use the same approach as send
			// In a full implementation, you'd decode and parse the raw message
			result, err := svc.DraftCreate(gmail.SendOptions{
				Body: raw,
			})
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&raw, "raw", "", "Base64-encoded RFC 2822 message (required)")
	cmd.MarkFlagRequired("raw")

	return cmd
}

func gmailAPIDraftSendCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "draft-send <draft-id>",
		Short: "Send a draft",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := gmailWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.DraftSend(args[0])
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}
}

func gmailAPIDraftUpdateCmd() *cobra.Command {
	var raw string

	cmd := &cobra.Command{
		Use:   "draft-update <draft-id>",
		Short: "Update a draft",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := gmailWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.DraftUpdate(args[0], gmail.SendOptions{
				Body: raw,
			})
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&raw, "raw", "", "Base64-encoded RFC 2822 message (required)")
	cmd.MarkFlagRequired("raw")

	return cmd
}

func gmailAPIDraftDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "draft-delete <draft-id>",
		Short: "Delete a draft",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := gmailWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			if err := svc.DraftDelete(args[0]); err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"draft_id": args[0],
				"deleted":  true,
			}, "write")
		},
	}
}
