// Package gmail provides Gmail API operations.
package gmail

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Service wraps the Gmail API service.
type Service struct {
	svc *gmail.Service
}

// NewService creates a new Gmail service.
func NewService(ctx context.Context, client *http.Client) (*Service, error) {
	svc, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}
	return &Service{svc: svc}, nil
}

// MessageSummary represents a summary of a Gmail message.
type MessageSummary struct {
	ID        string `json:"id"`
	ThreadID  string `json:"thread_id"`
	Subject   string `json:"subject"`
	From      string `json:"from"`
	To        string `json:"to,omitempty"`
	Date      string `json:"date"`
	Snippet   string `json:"snippet"`
	LabelIDs  []string `json:"label_ids,omitempty"`
	IsUnread  bool   `json:"is_unread"`
}

// MessageFull represents a full Gmail message.
type MessageFull struct {
	ID          string            `json:"id"`
	ThreadID    string            `json:"thread_id"`
	Subject     string            `json:"subject"`
	From        string            `json:"from"`
	To          []string          `json:"to"`
	Cc          []string          `json:"cc,omitempty"`
	Bcc         []string          `json:"bcc,omitempty"`
	Date        string            `json:"date"`
	BodyText    string            `json:"body_text"`
	BodyHTML    string            `json:"body_html,omitempty"`
	Attachments []AttachmentInfo  `json:"attachments"`
	LabelIDs    []string          `json:"label_ids"`
	Headers     map[string]string `json:"headers,omitempty"`
}

// AttachmentInfo represents information about a message attachment.
type AttachmentInfo struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
}

// ThreadSummary represents a summary of a Gmail thread.
type ThreadSummary struct {
	ID           string           `json:"id"`
	Subject      string           `json:"subject"`
	Snippet      string           `json:"snippet"`
	MessageCount int              `json:"message_count"`
	Messages     []MessageSummary `json:"messages"`
}

// LabelInfo represents information about a Gmail label.
type LabelInfo struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	Type                  string `json:"type"`
	MessagesTotal         int64  `json:"messages_total,omitempty"`
	MessagesUnread        int64  `json:"messages_unread,omitempty"`
	ThreadsTotal          int64  `json:"threads_total,omitempty"`
	ThreadsUnread         int64  `json:"threads_unread,omitempty"`
}

// SendResult represents the result of sending a message.
type SendResult struct {
	MessageID string `json:"message_id"`
	ThreadID  string `json:"thread_id"`
}

// DraftInfo represents information about a draft.
type DraftInfo struct {
	ID        string `json:"id"`
	MessageID string `json:"message_id,omitempty"`
	ThreadID  string `json:"thread_id,omitempty"`
	Subject   string `json:"subject,omitempty"`
	To        string `json:"to,omitempty"`
	Snippet   string `json:"snippet,omitempty"`
}
