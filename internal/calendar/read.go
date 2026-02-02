package calendar

import (
	"fmt"
	"time"

	"google.golang.org/api/calendar/v3"
)

// ListOptions contains options for listing events.
type ListOptions struct {
	CalendarID string
	TimeMin    time.Time
	TimeMax    time.Time
	Query      string
	MaxResults int64
	PageToken  string
}

// List returns events matching the criteria.
func (s *Service) List(opts ListOptions) ([]EventSummary, string, error) {
	calendarID := opts.CalendarID
	if calendarID == "" {
		calendarID = "primary"
	}

	call := s.svc.Events.List(calendarID).SingleEvents(true).OrderBy("startTime")

	if !opts.TimeMin.IsZero() {
		call = call.TimeMin(opts.TimeMin.Format(time.RFC3339))
	}
	if !opts.TimeMax.IsZero() {
		call = call.TimeMax(opts.TimeMax.Format(time.RFC3339))
	}
	if opts.Query != "" {
		call = call.Q(opts.Query)
	}
	if opts.MaxResults > 0 {
		call = call.MaxResults(opts.MaxResults)
	} else {
		call = call.MaxResults(10)
	}
	if opts.PageToken != "" {
		call = call.PageToken(opts.PageToken)
	}

	resp, err := call.Do()
	if err != nil {
		return nil, "", fmt.Errorf("failed to list events: %w", err)
	}

	events := make([]EventSummary, 0, len(resp.Items))
	for _, item := range resp.Items {
		events = append(events, parseEventToSummary(item))
	}

	return events, resp.NextPageToken, nil
}

// Today returns today's events.
func (s *Service) Today(calendarID string) ([]EventSummary, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	events, _, err := s.List(ListOptions{
		CalendarID: calendarID,
		TimeMin:    startOfDay,
		TimeMax:    endOfDay,
		MaxResults: 50,
	})
	return events, err
}

// Week returns this week's events.
func (s *Service) Week(calendarID string) ([]EventSummary, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfWeek := startOfDay.Add(7 * 24 * time.Hour)

	events, _, err := s.List(ListOptions{
		CalendarID: calendarID,
		TimeMin:    startOfDay,
		TimeMax:    endOfWeek,
		MaxResults: 100,
	})
	return events, err
}

// Upcoming returns events in the next N days.
func (s *Service) Upcoming(calendarID string, days int) ([]EventSummary, error) {
	now := time.Now()
	endTime := now.Add(time.Duration(days) * 24 * time.Hour)

	events, _, err := s.List(ListOptions{
		CalendarID: calendarID,
		TimeMin:    now,
		TimeMax:    endTime,
		MaxResults: 100,
	})
	return events, err
}

// Get returns a specific event.
func (s *Service) Get(calendarID, eventID string) (*EventFull, error) {
	if calendarID == "" {
		calendarID = "primary"
	}

	event, err := s.svc.Events.Get(calendarID, eventID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return parseEventToFull(event), nil
}

// Find searches for events matching the query.
func (s *Service) Find(query string, from, to time.Time) ([]EventSummary, error) {
	events, _, err := s.List(ListOptions{
		Query:      query,
		TimeMin:    from,
		TimeMax:    to,
		MaxResults: 50,
	})
	return events, err
}

// FreeBusy returns free/busy information for the given time range.
func (s *Service) FreeBusy(start, end time.Time, calendarIDs []string) (*FreeBusyResult, error) {
	if len(calendarIDs) == 0 {
		calendarIDs = []string{"primary"}
	}

	items := make([]*calendar.FreeBusyRequestItem, 0, len(calendarIDs))
	for _, id := range calendarIDs {
		items = append(items, &calendar.FreeBusyRequestItem{Id: id})
	}

	req := &calendar.FreeBusyRequest{
		TimeMin: start.Format(time.RFC3339),
		TimeMax: end.Format(time.RFC3339),
		Items:   items,
	}

	resp, err := s.svc.Freebusy.Query(req).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get free/busy info: %w", err)
	}

	// Aggregate busy slots from all calendars
	var busy []FreeBusySlot
	for _, cal := range resp.Calendars {
		for _, slot := range cal.Busy {
			busy = append(busy, FreeBusySlot{
				Start: slot.Start,
				End:   slot.End,
			})
		}
	}

	// Calculate free slots
	freeSlots := calculateFreeSlots(start, end, busy)

	return &FreeBusyResult{
		Busy:      busy,
		FreeSlots: freeSlots,
	}, nil
}

// Calendars returns a list of calendars.
func (s *Service) Calendars() ([]CalendarInfo, error) {
	resp, err := s.svc.CalendarList.List().Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list calendars: %w", err)
	}

	calendars := make([]CalendarInfo, 0, len(resp.Items))
	for _, item := range resp.Items {
		calendars = append(calendars, CalendarInfo{
			ID:          item.Id,
			Summary:     item.Summary,
			Description: item.Description,
			TimeZone:    item.TimeZone,
			Primary:     item.Primary,
		})
	}

	return calendars, nil
}

// parseEventToSummary converts a calendar event to an EventSummary.
func parseEventToSummary(event *calendar.Event) EventSummary {
	start, end, allDay := parseEventTimes(event)

	var attendees []string
	for _, att := range event.Attendees {
		attendees = append(attendees, att.Email)
	}

	var organizer string
	if event.Organizer != nil {
		organizer = event.Organizer.Email
	}

	return EventSummary{
		ID:          event.Id,
		Title:       event.Summary,
		Description: event.Description,
		Location:    event.Location,
		Start:       start,
		End:         end,
		AllDay:      allDay,
		Status:      event.Status,
		Attendees:   attendees,
		HTMLLink:    event.HtmlLink,
		Organizer:   organizer,
	}
}

// parseEventToFull converts a calendar event to an EventFull.
func parseEventToFull(event *calendar.Event) *EventFull {
	start, end, allDay := parseEventTimes(event)

	var attendees []AttendeeInfo
	for _, att := range event.Attendees {
		attendees = append(attendees, AttendeeInfo{
			Email:          att.Email,
			DisplayName:    att.DisplayName,
			ResponseStatus: att.ResponseStatus,
			Organizer:      att.Organizer,
			Self:           att.Self,
		})
	}

	var organizer string
	if event.Organizer != nil {
		organizer = event.Organizer.Email
	}

	var reminders *RemindersInfo
	if event.Reminders != nil {
		reminders = &RemindersInfo{
			UseDefault: event.Reminders.UseDefault,
		}
		for _, r := range event.Reminders.Overrides {
			reminders.Overrides = append(reminders.Overrides, ReminderInfo{
				Method:  r.Method,
				Minutes: r.Minutes,
			})
		}
	}

	var conference *ConferenceInfo
	if event.ConferenceData != nil {
		conference = &ConferenceInfo{
			ConferenceID: event.ConferenceData.ConferenceId,
		}
		for _, ep := range event.ConferenceData.EntryPoints {
			conference.EntryPoints = append(conference.EntryPoints, EntryPointInfo{
				Type: ep.EntryPointType,
				URI:  ep.Uri,
			})
		}
	}

	return &EventFull{
		ID:             event.Id,
		Title:          event.Summary,
		Description:    event.Description,
		Location:       event.Location,
		Start:          start,
		End:            end,
		AllDay:         allDay,
		Status:         event.Status,
		Attendees:      attendees,
		HTMLLink:       event.HtmlLink,
		Organizer:      organizer,
		Recurrence:     event.Recurrence,
		RecurringID:    event.RecurringEventId,
		Created:        event.Created,
		Updated:        event.Updated,
		Reminders:      reminders,
		ConferenceData: conference,
	}
}

// parseEventTimes extracts start/end times from an event.
func parseEventTimes(event *calendar.Event) (start, end string, allDay bool) {
	if event.Start == nil {
		return "", "", false
	}

	if event.Start.Date != "" {
		// All-day event
		return event.Start.Date, event.End.Date, true
	}

	return event.Start.DateTime, event.End.DateTime, false
}

// calculateFreeSlots calculates free time slots from busy slots.
func calculateFreeSlots(start, end time.Time, busy []FreeBusySlot) []FreeBusySlot {
	if len(busy) == 0 {
		return []FreeBusySlot{{
			Start: start.Format(time.RFC3339),
			End:   end.Format(time.RFC3339),
		}}
	}

	var freeSlots []FreeBusySlot
	currentStart := start

	for _, slot := range busy {
		busyStart, _ := time.Parse(time.RFC3339, slot.Start)
		busyEnd, _ := time.Parse(time.RFC3339, slot.End)

		if currentStart.Before(busyStart) {
			freeSlots = append(freeSlots, FreeBusySlot{
				Start: currentStart.Format(time.RFC3339),
				End:   busyStart.Format(time.RFC3339),
			})
		}

		if busyEnd.After(currentStart) {
			currentStart = busyEnd
		}
	}

	if currentStart.Before(end) {
		freeSlots = append(freeSlots, FreeBusySlot{
			Start: currentStart.Format(time.RFC3339),
			End:   end.Format(time.RFC3339),
		})
	}

	return freeSlots
}
