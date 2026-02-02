package gmail

import (
	"encoding/base64"
	"fmt"
	"strings"

	"google.golang.org/api/gmail/v1"
)

// ListOptions contains options for listing messages.
type ListOptions struct {
	LabelIDs   []string
	Query      string
	MaxResults int64
	PageToken  string
	UnreadOnly bool
}

// List returns a list of messages matching the criteria.
func (s *Service) List(opts ListOptions) ([]MessageSummary, string, error) {
	call := s.svc.Users.Messages.List("me")

	if len(opts.LabelIDs) > 0 {
		call = call.LabelIds(opts.LabelIDs...)
	}

	query := opts.Query
	if opts.UnreadOnly {
		if query != "" {
			query = query + " is:unread"
		} else {
			query = "is:unread"
		}
	}
	if query != "" {
		call = call.Q(query)
	}

	if opts.MaxResults > 0 {
		call = call.MaxResults(opts.MaxResults)
	} else {
		call = call.MaxResults(10) // Default
	}

	if opts.PageToken != "" {
		call = call.PageToken(opts.PageToken)
	}

	resp, err := call.Do()
	if err != nil {
		return nil, "", fmt.Errorf("failed to list messages: %w", err)
	}

	summaries := make([]MessageSummary, 0, len(resp.Messages))
	for _, msg := range resp.Messages {
		// Get message metadata for each message
		fullMsg, err := s.svc.Users.Messages.Get("me", msg.Id).Format("metadata").
			MetadataHeaders("From", "To", "Subject", "Date").Do()
		if err != nil {
			continue // Skip messages that fail to load
		}

		summary := parseMessageToSummary(fullMsg)
		summaries = append(summaries, summary)
	}

	return summaries, resp.NextPageToken, nil
}

// Inbox returns messages from the inbox.
func (s *Service) Inbox(limit int64, unreadOnly bool) ([]MessageSummary, error) {
	messages, _, err := s.List(ListOptions{
		LabelIDs:   []string{"INBOX"},
		MaxResults: limit,
		UnreadOnly: unreadOnly,
	})
	return messages, err
}

// Search searches for messages matching the query.
func (s *Service) Search(query string, limit int64) ([]MessageSummary, error) {
	messages, _, err := s.List(ListOptions{
		Query:      query,
		MaxResults: limit,
	})
	return messages, err
}

// Get returns a full message by ID.
func (s *Service) Get(messageID string) (*MessageFull, error) {
	msg, err := s.svc.Users.Messages.Get("me", messageID).Format("full").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return parseMessageToFull(msg), nil
}

// GetRaw returns a message in raw format (RFC 2822).
func (s *Service) GetRaw(messageID string) ([]byte, error) {
	msg, err := s.svc.Users.Messages.Get("me", messageID).Format("raw").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get raw message: %w", err)
	}

	raw, err := base64.URLEncoding.DecodeString(msg.Raw)
	if err != nil {
		return nil, fmt.Errorf("failed to decode raw message: %w", err)
	}

	return raw, nil
}

// GetThread returns all messages in a thread.
func (s *Service) GetThread(threadID string) (*ThreadSummary, error) {
	thread, err := s.svc.Users.Threads.Get("me", threadID).Format("metadata").
		MetadataHeaders("From", "To", "Subject", "Date").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get thread: %w", err)
	}

	messages := make([]MessageSummary, 0, len(thread.Messages))
	var subject string
	for i, msg := range thread.Messages {
		summary := parseMessageToSummary(msg)
		messages = append(messages, summary)
		if i == 0 {
			subject = summary.Subject
		}
	}

	return &ThreadSummary{
		ID:           thread.Id,
		Subject:      subject,
		Snippet:      thread.Snippet,
		MessageCount: len(messages),
		Messages:     messages,
	}, nil
}

// Labels returns all labels.
func (s *Service) Labels() ([]LabelInfo, error) {
	resp, err := s.svc.Users.Labels.List("me").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list labels: %w", err)
	}

	labels := make([]LabelInfo, 0, len(resp.Labels))
	for _, label := range resp.Labels {
		labels = append(labels, LabelInfo{
			ID:             label.Id,
			Name:           label.Name,
			Type:           label.Type,
			MessagesTotal:  label.MessagesTotal,
			MessagesUnread: label.MessagesUnread,
			ThreadsTotal:   label.ThreadsTotal,
			ThreadsUnread:  label.ThreadsUnread,
		})
	}

	return labels, nil
}

// GetAttachment retrieves an attachment from a message.
func (s *Service) GetAttachment(messageID, attachmentID string) ([]byte, error) {
	att, err := s.svc.Users.Messages.Attachments.Get("me", messageID, attachmentID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get attachment: %w", err)
	}

	data, err := base64.URLEncoding.DecodeString(att.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode attachment: %w", err)
	}

	return data, nil
}

// parseMessageToSummary converts a Gmail message to a MessageSummary.
func parseMessageToSummary(msg *gmail.Message) MessageSummary {
	headers := extractHeaders(msg.Payload)

	isUnread := false
	for _, labelID := range msg.LabelIds {
		if labelID == "UNREAD" {
			isUnread = true
			break
		}
	}

	return MessageSummary{
		ID:       msg.Id,
		ThreadID: msg.ThreadId,
		Subject:  headers["Subject"],
		From:     headers["From"],
		To:       headers["To"],
		Date:     headers["Date"],
		Snippet:  msg.Snippet,
		LabelIDs: msg.LabelIds,
		IsUnread: isUnread,
	}
}

// parseMessageToFull converts a Gmail message to a MessageFull.
func parseMessageToFull(msg *gmail.Message) *MessageFull {
	headers := extractHeaders(msg.Payload)
	bodyText, bodyHTML := extractBody(msg.Payload)
	attachments := extractAttachments(msg.Payload)

	// Parse To, Cc, Bcc addresses
	parseAddresses := func(s string) []string {
		if s == "" {
			return nil
		}
		parts := strings.Split(s, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			result = append(result, strings.TrimSpace(p))
		}
		return result
	}

	return &MessageFull{
		ID:          msg.Id,
		ThreadID:    msg.ThreadId,
		Subject:     headers["Subject"],
		From:        headers["From"],
		To:          parseAddresses(headers["To"]),
		Cc:          parseAddresses(headers["Cc"]),
		Bcc:         parseAddresses(headers["Bcc"]),
		Date:        headers["Date"],
		BodyText:    bodyText,
		BodyHTML:    bodyHTML,
		Attachments: attachments,
		LabelIDs:    msg.LabelIds,
		Headers:     headers,
	}
}

// extractHeaders extracts headers from a message payload.
func extractHeaders(payload *gmail.MessagePart) map[string]string {
	headers := make(map[string]string)
	if payload == nil {
		return headers
	}

	for _, h := range payload.Headers {
		headers[h.Name] = h.Value
	}
	return headers
}

// extractBody extracts the plain text and HTML body from a message payload.
func extractBody(payload *gmail.MessagePart) (text, html string) {
	if payload == nil {
		return "", ""
	}

	// Check if this part has a body
	if payload.Body != nil && payload.Body.Data != "" {
		decoded, err := base64.URLEncoding.DecodeString(payload.Body.Data)
		if err == nil {
			if payload.MimeType == "text/plain" {
				text = string(decoded)
			} else if payload.MimeType == "text/html" {
				html = string(decoded)
			}
		}
	}

	// Recursively check parts
	for _, part := range payload.Parts {
		partText, partHTML := extractBody(part)
		if text == "" && partText != "" {
			text = partText
		}
		if html == "" && partHTML != "" {
			html = partHTML
		}
	}

	return text, html
}

// extractAttachments extracts attachment information from a message payload.
func extractAttachments(payload *gmail.MessagePart) []AttachmentInfo {
	var attachments []AttachmentInfo

	var extract func(part *gmail.MessagePart)
	extract = func(part *gmail.MessagePart) {
		if part == nil {
			return
		}

		if part.Filename != "" && part.Body != nil && part.Body.AttachmentId != "" {
			attachments = append(attachments, AttachmentInfo{
				ID:       part.Body.AttachmentId,
				Filename: part.Filename,
				MimeType: part.MimeType,
				Size:     part.Body.Size,
			})
		}

		for _, p := range part.Parts {
			extract(p)
		}
	}

	extract(payload)
	return attachments
}
