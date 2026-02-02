package gmail

import (
	"encoding/base64"
	"fmt"
	"strings"

	"google.golang.org/api/gmail/v1"
)

// SendOptions contains options for sending a message.
type SendOptions struct {
	To      []string
	Cc      []string
	Bcc     []string
	Subject string
	Body    string
}

// Send sends an email message.
func (s *Service) Send(opts SendOptions) (*SendResult, error) {
	raw := buildRawMessage(opts.To, opts.Cc, opts.Bcc, opts.Subject, opts.Body, nil)

	msg := &gmail.Message{
		Raw: base64.URLEncoding.EncodeToString([]byte(raw)),
	}

	sent, err := s.svc.Users.Messages.Send("me", msg).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return &SendResult{
		MessageID: sent.Id,
		ThreadID:  sent.ThreadId,
	}, nil
}

// ReplyOptions contains options for replying to a message.
type ReplyOptions struct {
	MessageID string
	Body      string
	ReplyAll  bool
}

// Reply sends a reply to an existing message.
func (s *Service) Reply(opts ReplyOptions) (*SendResult, error) {
	// Get the original message
	original, err := s.Get(opts.MessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get original message: %w", err)
	}

	// Build recipients
	to := []string{original.From}
	var cc []string
	if opts.ReplyAll {
		// Add original recipients (excluding self) to Cc
		for _, addr := range original.To {
			if !strings.Contains(strings.ToLower(addr), "me") {
				cc = append(cc, addr)
			}
		}
		cc = append(cc, original.Cc...)
	}

	// Build subject with Re: prefix
	subject := original.Subject
	if !strings.HasPrefix(strings.ToLower(subject), "re:") {
		subject = "Re: " + subject
	}

	// Build headers for threading
	headers := map[string]string{
		"In-Reply-To": original.Headers["Message-ID"],
		"References":  original.Headers["References"] + " " + original.Headers["Message-ID"],
	}

	raw := buildRawMessage(to, cc, nil, subject, opts.Body, headers)

	msg := &gmail.Message{
		Raw:      base64.URLEncoding.EncodeToString([]byte(raw)),
		ThreadId: original.ThreadID,
	}

	sent, err := s.svc.Users.Messages.Send("me", msg).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to send reply: %w", err)
	}

	return &SendResult{
		MessageID: sent.Id,
		ThreadID:  sent.ThreadId,
	}, nil
}

// ForwardOptions contains options for forwarding a message.
type ForwardOptions struct {
	MessageID string
	To        []string
	Body      string // Optional additional message
}

// Forward forwards a message to new recipients.
func (s *Service) Forward(opts ForwardOptions) (*SendResult, error) {
	// Get the original message
	original, err := s.Get(opts.MessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get original message: %w", err)
	}

	// Build subject with Fwd: prefix
	subject := original.Subject
	if !strings.HasPrefix(strings.ToLower(subject), "fwd:") {
		subject = "Fwd: " + subject
	}

	// Build body with optional message and quoted original
	body := opts.Body
	if body != "" {
		body += "\n\n"
	}
	body += "---------- Forwarded message ---------\n"
	body += fmt.Sprintf("From: %s\n", original.From)
	body += fmt.Sprintf("Date: %s\n", original.Date)
	body += fmt.Sprintf("Subject: %s\n", original.Subject)
	body += fmt.Sprintf("To: %s\n", strings.Join(original.To, ", "))
	body += "\n" + original.BodyText

	raw := buildRawMessage(opts.To, nil, nil, subject, body, nil)

	msg := &gmail.Message{
		Raw: base64.URLEncoding.EncodeToString([]byte(raw)),
	}

	sent, err := s.svc.Users.Messages.Send("me", msg).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to forward message: %w", err)
	}

	return &SendResult{
		MessageID: sent.Id,
		ThreadID:  sent.ThreadId,
	}, nil
}

// DraftCreate creates a draft message.
func (s *Service) DraftCreate(opts SendOptions) (*DraftInfo, error) {
	raw := buildRawMessage(opts.To, opts.Cc, opts.Bcc, opts.Subject, opts.Body, nil)

	msg := &gmail.Message{
		Raw: base64.URLEncoding.EncodeToString([]byte(raw)),
	}

	draft := &gmail.Draft{
		Message: msg,
	}

	created, err := s.svc.Users.Drafts.Create("me", draft).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create draft: %w", err)
	}

	return &DraftInfo{
		ID:        created.Id,
		MessageID: created.Message.Id,
	}, nil
}

// DraftSend sends a draft.
func (s *Service) DraftSend(draftID string) (*SendResult, error) {
	draft := &gmail.Draft{
		Id: draftID,
	}

	sent, err := s.svc.Users.Drafts.Send("me", draft).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to send draft: %w", err)
	}

	return &SendResult{
		MessageID: sent.Id,
		ThreadID:  sent.ThreadId,
	}, nil
}

// DraftUpdate updates an existing draft.
func (s *Service) DraftUpdate(draftID string, opts SendOptions) (*DraftInfo, error) {
	raw := buildRawMessage(opts.To, opts.Cc, opts.Bcc, opts.Subject, opts.Body, nil)

	msg := &gmail.Message{
		Raw: base64.URLEncoding.EncodeToString([]byte(raw)),
	}

	draft := &gmail.Draft{
		Message: msg,
	}

	updated, err := s.svc.Users.Drafts.Update("me", draftID, draft).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update draft: %w", err)
	}

	return &DraftInfo{
		ID:        updated.Id,
		MessageID: updated.Message.Id,
	}, nil
}

// DraftDelete deletes a draft.
func (s *Service) DraftDelete(draftID string) error {
	if err := s.svc.Users.Drafts.Delete("me", draftID).Do(); err != nil {
		return fmt.Errorf("failed to delete draft: %w", err)
	}
	return nil
}

// DraftList returns all drafts.
func (s *Service) DraftList(maxResults int64) ([]DraftInfo, error) {
	call := s.svc.Users.Drafts.List("me")
	if maxResults > 0 {
		call = call.MaxResults(maxResults)
	}

	resp, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list drafts: %w", err)
	}

	drafts := make([]DraftInfo, 0, len(resp.Drafts))
	for _, d := range resp.Drafts {
		drafts = append(drafts, DraftInfo{
			ID:        d.Id,
			MessageID: d.Message.Id,
		})
	}

	return drafts, nil
}

// ModifyLabels modifies the labels on a message.
func (s *Service) ModifyLabels(messageID string, addLabels, removeLabels []string) error {
	req := &gmail.ModifyMessageRequest{
		AddLabelIds:    addLabels,
		RemoveLabelIds: removeLabels,
	}

	_, err := s.svc.Users.Messages.Modify("me", messageID, req).Do()
	if err != nil {
		return fmt.Errorf("failed to modify labels: %w", err)
	}

	return nil
}

// Trash moves a message to trash.
func (s *Service) Trash(messageID string) error {
	_, err := s.svc.Users.Messages.Trash("me", messageID).Do()
	if err != nil {
		return fmt.Errorf("failed to trash message: %w", err)
	}
	return nil
}

// Untrash removes a message from trash.
func (s *Service) Untrash(messageID string) error {
	_, err := s.svc.Users.Messages.Untrash("me", messageID).Do()
	if err != nil {
		return fmt.Errorf("failed to untrash message: %w", err)
	}
	return nil
}

// Delete permanently deletes a message.
func (s *Service) Delete(messageID string) error {
	if err := s.svc.Users.Messages.Delete("me", messageID).Do(); err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}
	return nil
}

// SendRaw sends a raw RFC 2822 message.
func (s *Service) SendRaw(raw []byte) (*SendResult, error) {
	msg := &gmail.Message{
		Raw: base64.URLEncoding.EncodeToString(raw),
	}

	sent, err := s.svc.Users.Messages.Send("me", msg).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to send raw message: %w", err)
	}

	return &SendResult{
		MessageID: sent.Id,
		ThreadID:  sent.ThreadId,
	}, nil
}

// buildRawMessage builds a raw RFC 2822 message.
func buildRawMessage(to, cc, bcc []string, subject, body string, extraHeaders map[string]string) string {
	var msg strings.Builder

	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
	if len(cc) > 0 {
		msg.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(cc, ", ")))
	}
	if len(bcc) > 0 {
		msg.WriteString(fmt.Sprintf("Bcc: %s\r\n", strings.Join(bcc, ", ")))
	}
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")

	for key, value := range extraHeaders {
		if value != "" {
			msg.WriteString(fmt.Sprintf("%s: %s\r\n", key, strings.TrimSpace(value)))
		}
	}

	msg.WriteString("\r\n")
	msg.WriteString(body)

	return msg.String()
}
