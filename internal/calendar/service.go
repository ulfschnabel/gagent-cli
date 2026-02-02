// Package calendar provides Google Calendar API operations.
package calendar

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Service wraps the Calendar API service.
type Service struct {
	svc *calendar.Service
}

// NewService creates a new Calendar service.
func NewService(ctx context.Context, client *http.Client) (*Service, error) {
	svc, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Calendar service: %w", err)
	}
	return &Service{svc: svc}, nil
}

// EventSummary represents a summary of a calendar event.
type EventSummary struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Location    string   `json:"location,omitempty"`
	Start       string   `json:"start"`
	End         string   `json:"end"`
	AllDay      bool     `json:"all_day"`
	Status      string   `json:"status"`
	Attendees   []string `json:"attendees,omitempty"`
	HTMLLink    string   `json:"html_link"`
	Organizer   string   `json:"organizer,omitempty"`
}

// EventFull represents a full calendar event.
type EventFull struct {
	ID              string           `json:"id"`
	Title           string           `json:"title"`
	Description     string           `json:"description,omitempty"`
	Location        string           `json:"location,omitempty"`
	Start           string           `json:"start"`
	End             string           `json:"end"`
	AllDay          bool             `json:"all_day"`
	Status          string           `json:"status"`
	Attendees       []AttendeeInfo   `json:"attendees,omitempty"`
	HTMLLink        string           `json:"html_link"`
	Organizer       string           `json:"organizer,omitempty"`
	Recurrence      []string         `json:"recurrence,omitempty"`
	RecurringID     string           `json:"recurring_event_id,omitempty"`
	Created         string           `json:"created"`
	Updated         string           `json:"updated"`
	Reminders       *RemindersInfo   `json:"reminders,omitempty"`
	ConferenceData  *ConferenceInfo  `json:"conference_data,omitempty"`
}

// AttendeeInfo represents information about an event attendee.
type AttendeeInfo struct {
	Email          string `json:"email"`
	DisplayName    string `json:"display_name,omitempty"`
	ResponseStatus string `json:"response_status"`
	Organizer      bool   `json:"organizer,omitempty"`
	Self           bool   `json:"self,omitempty"`
}

// RemindersInfo represents reminder settings.
type RemindersInfo struct {
	UseDefault bool             `json:"use_default"`
	Overrides  []ReminderInfo   `json:"overrides,omitempty"`
}

// ReminderInfo represents a single reminder.
type ReminderInfo struct {
	Method  string `json:"method"`
	Minutes int64  `json:"minutes"`
}

// ConferenceInfo represents conference/meeting data.
type ConferenceInfo struct {
	EntryPoints []EntryPointInfo `json:"entry_points,omitempty"`
	ConferenceID string          `json:"conference_id,omitempty"`
}

// EntryPointInfo represents a conference entry point.
type EntryPointInfo struct {
	Type string `json:"type"`
	URI  string `json:"uri"`
}

// CalendarInfo represents information about a calendar.
type CalendarInfo struct {
	ID          string `json:"id"`
	Summary     string `json:"summary"`
	Description string `json:"description,omitempty"`
	TimeZone    string `json:"time_zone,omitempty"`
	Primary     bool   `json:"primary,omitempty"`
}

// FreeBusySlot represents a busy time slot.
type FreeBusySlot struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// FreeBusyResult represents the result of a free/busy query.
type FreeBusyResult struct {
	Busy      []FreeBusySlot `json:"busy"`
	FreeSlots []FreeBusySlot `json:"free_slots,omitempty"`
}

// CreateEventResult represents the result of creating an event.
type CreateEventResult struct {
	EventID  string `json:"event_id"`
	HTMLLink string `json:"html_link"`
}
