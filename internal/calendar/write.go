package calendar

import (
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/api/calendar/v3"
)

// CreateEventOptions contains options for creating an event.
type CreateEventOptions struct {
	CalendarID  string
	Title       string
	Description string
	Location    string
	Start       time.Time
	End         time.Time
	AllDay      bool
	Attendees   []string
}

// Create creates a new calendar event.
func (s *Service) Create(opts CreateEventOptions) (*CreateEventResult, error) {
	calendarID := opts.CalendarID
	if calendarID == "" {
		calendarID = "primary"
	}

	event := &calendar.Event{
		Summary:     opts.Title,
		Description: opts.Description,
		Location:    opts.Location,
	}

	if opts.AllDay {
		event.Start = &calendar.EventDateTime{
			Date: opts.Start.Format("2006-01-02"),
		}
		event.End = &calendar.EventDateTime{
			Date: opts.End.Format("2006-01-02"),
		}
	} else {
		event.Start = &calendar.EventDateTime{
			DateTime: opts.Start.Format(time.RFC3339),
		}
		event.End = &calendar.EventDateTime{
			DateTime: opts.End.Format(time.RFC3339),
		}
	}

	if len(opts.Attendees) > 0 {
		for _, email := range opts.Attendees {
			event.Attendees = append(event.Attendees, &calendar.EventAttendee{
				Email: email,
			})
		}
	}

	created, err := s.svc.Events.Insert(calendarID, event).SendUpdates("all").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	return &CreateEventResult{
		EventID:  created.Id,
		HTMLLink: created.HtmlLink,
	}, nil
}

// RescheduleOptions contains options for rescheduling an event.
type RescheduleOptions struct {
	CalendarID string
	EventID    string
	Start      time.Time
	End        time.Time
}

// Reschedule updates an event's time.
func (s *Service) Reschedule(opts RescheduleOptions) (*CreateEventResult, error) {
	calendarID := opts.CalendarID
	if calendarID == "" {
		calendarID = "primary"
	}

	// Get the existing event
	event, err := s.svc.Events.Get(calendarID, opts.EventID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	// Update times
	if event.Start.Date != "" {
		// All-day event
		event.Start = &calendar.EventDateTime{
			Date: opts.Start.Format("2006-01-02"),
		}
		event.End = &calendar.EventDateTime{
			Date: opts.End.Format("2006-01-02"),
		}
	} else {
		event.Start = &calendar.EventDateTime{
			DateTime: opts.Start.Format(time.RFC3339),
		}
		event.End = &calendar.EventDateTime{
			DateTime: opts.End.Format(time.RFC3339),
		}
	}

	updated, err := s.svc.Events.Update(calendarID, opts.EventID, event).SendUpdates("all").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to reschedule event: %w", err)
	}

	return &CreateEventResult{
		EventID:  updated.Id,
		HTMLLink: updated.HtmlLink,
	}, nil
}

// Cancel deletes an event.
func (s *Service) Cancel(calendarID, eventID string, notify bool) error {
	if calendarID == "" {
		calendarID = "primary"
	}

	sendUpdates := "none"
	if notify {
		sendUpdates = "all"
	}

	if err := s.svc.Events.Delete(calendarID, eventID).SendUpdates(sendUpdates).Do(); err != nil {
		return fmt.Errorf("failed to cancel event: %w", err)
	}

	return nil
}

// Respond updates the response status for an event invitation.
func (s *Service) Respond(calendarID, eventID, status string) error {
	if calendarID == "" {
		calendarID = "primary"
	}

	// Valid statuses: accepted, declined, tentative
	validStatuses := map[string]bool{
		"accepted":  true,
		"declined":  true,
		"tentative": true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s (use: accepted, declined, tentative)", status)
	}

	// Get the event
	event, err := s.svc.Events.Get(calendarID, eventID).Do()
	if err != nil {
		return fmt.Errorf("failed to get event: %w", err)
	}

	// Find self in attendees and update status
	for _, att := range event.Attendees {
		if att.Self {
			att.ResponseStatus = status
			break
		}
	}

	_, err = s.svc.Events.Update(calendarID, eventID, event).Do()
	if err != nil {
		return fmt.Errorf("failed to respond to event: %w", err)
	}

	return nil
}

// QuickAdd creates an event using natural language.
func (s *Service) QuickAdd(calendarID, text string) (*CreateEventResult, error) {
	if calendarID == "" {
		calendarID = "primary"
	}

	event, err := s.svc.Events.QuickAdd(calendarID, text).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to quick add event: %w", err)
	}

	return &CreateEventResult{
		EventID:  event.Id,
		HTMLLink: event.HtmlLink,
	}, nil
}

// InsertRaw inserts an event from JSON.
func (s *Service) InsertRaw(calendarID string, eventJSON string) (*CreateEventResult, error) {
	if calendarID == "" {
		calendarID = "primary"
	}

	var event calendar.Event
	if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
		return nil, fmt.Errorf("failed to parse event JSON: %w", err)
	}

	created, err := s.svc.Events.Insert(calendarID, &event).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to insert event: %w", err)
	}

	return &CreateEventResult{
		EventID:  created.Id,
		HTMLLink: created.HtmlLink,
	}, nil
}

// UpdateRaw updates an event from JSON.
func (s *Service) UpdateRaw(calendarID, eventID, eventJSON string) (*CreateEventResult, error) {
	if calendarID == "" {
		calendarID = "primary"
	}

	var event calendar.Event
	if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
		return nil, fmt.Errorf("failed to parse event JSON: %w", err)
	}

	updated, err := s.svc.Events.Update(calendarID, eventID, &event).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	return &CreateEventResult{
		EventID:  updated.Id,
		HTMLLink: updated.HtmlLink,
	}, nil
}

// PatchRaw patches an event from JSON.
func (s *Service) PatchRaw(calendarID, eventID, patchJSON string) (*CreateEventResult, error) {
	if calendarID == "" {
		calendarID = "primary"
	}

	var event calendar.Event
	if err := json.Unmarshal([]byte(patchJSON), &event); err != nil {
		return nil, fmt.Errorf("failed to parse patch JSON: %w", err)
	}

	updated, err := s.svc.Events.Patch(calendarID, eventID, &event).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to patch event: %w", err)
	}

	return &CreateEventResult{
		EventID:  updated.Id,
		HTMLLink: updated.HtmlLink,
	}, nil
}

// Delete deletes an event with configurable notification.
func (s *Service) Delete(calendarID, eventID, sendUpdates string) error {
	if calendarID == "" {
		calendarID = "primary"
	}
	if sendUpdates == "" {
		sendUpdates = "all"
	}

	if err := s.svc.Events.Delete(calendarID, eventID).SendUpdates(sendUpdates).Do(); err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	return nil
}
