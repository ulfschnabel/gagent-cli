package main

import (
	"context"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/ulfhaga/gagent-cli/internal/auth"
	"github.com/ulfhaga/gagent-cli/internal/calendar"
	"github.com/ulfhaga/gagent-cli/internal/config"
	"github.com/ulfhaga/gagent-cli/internal/output"
)

// calendarReadService creates a Calendar service with read scope.
func calendarReadService(ctx context.Context) (*calendar.Service, error) {
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

	return calendar.NewService(ctx, client)
}

// calendarWriteService creates a Calendar service with write scope.
func calendarWriteService(ctx context.Context) (*calendar.Service, error) {
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

	return calendar.NewService(ctx, client)
}

func calendarTodayCmd() *cobra.Command {
	var calendarID string

	cmd := &cobra.Command{
		Use:   "today",
		Short: "List today's events",
		Long:  "Returns today's events with title, time, location, attendees.",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := calendarReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			events, err := svc.Today(calendarID)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"events": events,
				"count":  len(events),
				"date":   time.Now().Format("2006-01-02"),
			}, "read")
		},
	}

	cmd.Flags().StringVar(&calendarID, "calendar", "", "Calendar ID (default: primary)")

	return cmd
}

func calendarWeekCmd() *cobra.Command {
	var calendarID string

	cmd := &cobra.Command{
		Use:   "week",
		Short: "List this week's events",
		Long:  "Returns this week's events.",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := calendarReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			events, err := svc.Week(calendarID)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"events": events,
				"count":  len(events),
			}, "read")
		},
	}

	cmd.Flags().StringVar(&calendarID, "calendar", "", "Calendar ID (default: primary)")

	return cmd
}

func calendarUpcomingCmd() *cobra.Command {
	var calendarID string
	var days int

	cmd := &cobra.Command{
		Use:   "upcoming",
		Short: "List upcoming events",
		Long:  "Returns events in next N days (default 7).",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := calendarReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			events, err := svc.Upcoming(calendarID, days)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"events": events,
				"count":  len(events),
				"days":   days,
			}, "read")
		},
	}

	cmd.Flags().StringVar(&calendarID, "calendar", "", "Calendar ID (default: primary)")
	cmd.Flags().IntVar(&days, "days", 7, "Number of days to look ahead")

	return cmd
}

func calendarEventCmd() *cobra.Command {
	var calendarID string

	cmd := &cobra.Command{
		Use:   "event <event-id>",
		Short: "Get event details",
		Long:  "Returns full event details.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := calendarReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			event, err := svc.Get(calendarID, args[0])
			if err != nil {
				output.NotFoundError("Event", args[0])
				return
			}

			output.Success(event, "read")
		},
	}

	cmd.Flags().StringVar(&calendarID, "calendar", "", "Calendar ID (default: primary)")

	return cmd
}

func calendarFindCmd() *cobra.Command {
	var from, to string

	cmd := &cobra.Command{
		Use:   "find <query>",
		Short: "Search events",
		Long:  "Search events by text.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := calendarReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			var fromTime, toTime time.Time
			if from != "" {
				fromTime, _ = time.Parse(time.RFC3339, from)
			} else {
				fromTime = time.Now().AddDate(-1, 0, 0) // Default: 1 year ago
			}
			if to != "" {
				toTime, _ = time.Parse(time.RFC3339, to)
			} else {
				toTime = time.Now().AddDate(1, 0, 0) // Default: 1 year from now
			}

			events, err := svc.Find(args[0], fromTime, toTime)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"query":  args[0],
				"events": events,
				"count":  len(events),
			}, "read")
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "Start date (RFC3339)")
	cmd.Flags().StringVar(&to, "to", "", "End date (RFC3339)")

	return cmd
}

func calendarFreeBusyCmd() *cobra.Command {
	var start, end, calendars string

	cmd := &cobra.Command{
		Use:   "free-busy",
		Short: "Check availability",
		Long:  "Returns busy/free slots in time range.",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := calendarReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			startTime, err := time.Parse(time.RFC3339, start)
			if err != nil {
				// Try simpler format
				startTime, err = time.Parse("2006-01-02T15:04:05", start)
				if err != nil {
					output.InvalidInputError("Invalid start time format. Use RFC3339 (e.g., 2024-01-16T09:00:00Z)")
					return
				}
			}

			endTime, err := time.Parse(time.RFC3339, end)
			if err != nil {
				endTime, err = time.Parse("2006-01-02T15:04:05", end)
				if err != nil {
					output.InvalidInputError("Invalid end time format. Use RFC3339 (e.g., 2024-01-16T17:00:00Z)")
					return
				}
			}

			var calendarIDs []string
			if calendars != "" {
				calendarIDs = strings.Split(calendars, ",")
			}

			result, err := svc.FreeBusy(startTime, endTime, calendarIDs)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "read")
		},
	}

	cmd.Flags().StringVar(&start, "start", "", "Start datetime (required)")
	cmd.Flags().StringVar(&end, "end", "", "End datetime (required)")
	cmd.Flags().StringVar(&calendars, "calendars", "", "Calendar IDs (comma-separated)")

	cmd.MarkFlagRequired("start")
	cmd.MarkFlagRequired("end")

	return cmd
}

func calendarScheduleCmd() *cobra.Command {
	var title, description, location, calendarID, attendees string
	var start, end string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "schedule",
		Short: "Create an event",
		Long:  "Creates event, sends invites if attendees specified.",
		Run: func(cmd *cobra.Command, args []string) {
			startTime, err := time.Parse(time.RFC3339, start)
			if err != nil {
				startTime, err = time.Parse("2006-01-02T15:04:05", start)
				if err != nil {
					output.InvalidInputError("Invalid start time format")
					return
				}
			}

			endTime, err := time.Parse(time.RFC3339, end)
			if err != nil {
				endTime, err = time.Parse("2006-01-02T15:04:05", end)
				if err != nil {
					output.InvalidInputError("Invalid end time format")
					return
				}
			}

			var attendeeList []string
			if attendees != "" {
				attendeeList = strings.Split(attendees, ",")
			}

			if dryRun {
				output.SuccessNoScope(map[string]interface{}{
					"dry_run":     true,
					"title":       title,
					"start":       startTime.Format(time.RFC3339),
					"end":         endTime.Format(time.RFC3339),
					"location":    location,
					"description": description,
					"attendees":   attendeeList,
				})
				return
			}

			ctx := context.Background()
			svc, err := calendarWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.Create(calendar.CreateEventOptions{
				CalendarID:  calendarID,
				Title:       title,
				Description: description,
				Location:    location,
				Start:       startTime,
				End:         endTime,
				Attendees:   attendeeList,
			})
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Event title (required)")
	cmd.Flags().StringVar(&start, "start", "", "Start datetime (required)")
	cmd.Flags().StringVar(&end, "end", "", "End datetime (required)")
	cmd.Flags().StringVar(&calendarID, "calendar", "", "Calendar ID (default: primary)")
	cmd.Flags().StringVar(&location, "location", "", "Event location")
	cmd.Flags().StringVar(&description, "description", "", "Event description")
	cmd.Flags().StringVar(&attendees, "attendees", "", "Attendee emails (comma-separated)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be created")

	cmd.MarkFlagRequired("title")
	cmd.MarkFlagRequired("start")
	cmd.MarkFlagRequired("end")

	return cmd
}

func calendarRescheduleCmd() *cobra.Command {
	var start, end, calendarID string

	cmd := &cobra.Command{
		Use:   "reschedule <event-id>",
		Short: "Reschedule an event",
		Long:  "Updates event time, notifies attendees.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			startTime, err := time.Parse(time.RFC3339, start)
			if err != nil {
				startTime, err = time.Parse("2006-01-02T15:04:05", start)
				if err != nil {
					output.InvalidInputError("Invalid start time format")
					return
				}
			}

			endTime, err := time.Parse(time.RFC3339, end)
			if err != nil {
				endTime, err = time.Parse("2006-01-02T15:04:05", end)
				if err != nil {
					output.InvalidInputError("Invalid end time format")
					return
				}
			}

			ctx := context.Background()
			svc, err := calendarWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.Reschedule(calendar.RescheduleOptions{
				CalendarID: calendarID,
				EventID:    args[0],
				Start:      startTime,
				End:        endTime,
			})
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&start, "start", "", "New start datetime (required)")
	cmd.Flags().StringVar(&end, "end", "", "New end datetime (required)")
	cmd.Flags().StringVar(&calendarID, "calendar", "", "Calendar ID (default: primary)")

	cmd.MarkFlagRequired("start")
	cmd.MarkFlagRequired("end")

	return cmd
}

func calendarCancelCmd() *cobra.Command {
	var calendarID string
	var notify bool

	cmd := &cobra.Command{
		Use:   "cancel <event-id>",
		Short: "Cancel an event",
		Long:  "Deletes event, optionally notifies attendees.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := calendarWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			if err := svc.Cancel(calendarID, args[0], notify); err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"event_id":  args[0],
				"cancelled": true,
				"notified":  notify,
			}, "write")
		},
	}

	cmd.Flags().StringVar(&calendarID, "calendar", "", "Calendar ID (default: primary)")
	cmd.Flags().BoolVar(&notify, "notify", true, "Notify attendees")

	return cmd
}

func calendarRespondCmd() *cobra.Command {
	var calendarID, status string

	cmd := &cobra.Command{
		Use:   "respond <event-id>",
		Short: "Respond to invitation",
		Long:  "Responds to calendar invitation (accepted/declined/tentative).",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := calendarWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			if err := svc.Respond(calendarID, args[0], status); err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"event_id": args[0],
				"status":   status,
			}, "write")
		},
	}

	cmd.Flags().StringVar(&calendarID, "calendar", "", "Calendar ID (default: primary)")
	cmd.Flags().StringVar(&status, "status", "", "Response: accepted, declined, tentative (required)")

	cmd.MarkFlagRequired("status")

	return cmd
}

// Calendar API commands
func calendarAPICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Low-level Calendar API commands",
		Long:  "Direct access to Calendar API operations.",
	}

	cmd.AddCommand(calendarAPICalendarsCmd())
	cmd.AddCommand(calendarAPIEventsCmd())
	cmd.AddCommand(calendarAPIGetCmd())
	cmd.AddCommand(calendarAPIInsertCmd())
	cmd.AddCommand(calendarAPIUpdateCmd())
	cmd.AddCommand(calendarAPIPatchCmd())
	cmd.AddCommand(calendarAPIDeleteCmd())
	cmd.AddCommand(calendarAPIQuickAddCmd())

	return cmd
}

func calendarAPICalendarsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "calendars",
		Short: "List calendars",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := calendarReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			calendars, err := svc.Calendars()
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"calendars": calendars,
				"count":     len(calendars),
			}, "read")
		},
	}
}

func calendarAPIEventsCmd() *cobra.Command {
	var calendarID, timeMin, timeMax, pageToken string
	var limit int64

	cmd := &cobra.Command{
		Use:   "events",
		Short: "List events",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := calendarReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			opts := calendar.ListOptions{
				CalendarID: calendarID,
				MaxResults: limit,
				PageToken:  pageToken,
			}

			if timeMin != "" {
				t, _ := time.Parse(time.RFC3339, timeMin)
				opts.TimeMin = t
			}
			if timeMax != "" {
				t, _ := time.Parse(time.RFC3339, timeMax)
				opts.TimeMax = t
			}

			events, nextPageToken, err := svc.List(opts)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"events":          events,
				"count":           len(events),
				"next_page_token": nextPageToken,
			}, "read")
		},
	}

	cmd.Flags().StringVar(&calendarID, "calendar", "", "Calendar ID (default: primary)")
	cmd.Flags().StringVar(&timeMin, "time-min", "", "Minimum time (RFC3339)")
	cmd.Flags().StringVar(&timeMax, "time-max", "", "Maximum time (RFC3339)")
	cmd.Flags().Int64VarP(&limit, "limit", "n", 10, "Maximum events")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Page token")

	return cmd
}

func calendarAPIGetCmd() *cobra.Command {
	var calendarID string

	cmd := &cobra.Command{
		Use:   "get <event-id>",
		Short: "Get an event",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := calendarReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			event, err := svc.Get(calendarID, args[0])
			if err != nil {
				output.NotFoundError("Event", args[0])
				return
			}

			output.Success(event, "read")
		},
	}

	cmd.Flags().StringVar(&calendarID, "calendar", "", "Calendar ID (default: primary)")

	return cmd
}

func calendarAPIInsertCmd() *cobra.Command {
	var calendarID, eventJSON string

	cmd := &cobra.Command{
		Use:   "insert",
		Short: "Insert an event from JSON",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := calendarWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.InsertRaw(calendarID, eventJSON)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&calendarID, "calendar", "", "Calendar ID (default: primary)")
	cmd.Flags().StringVar(&eventJSON, "event-json", "", "Event JSON (required)")

	cmd.MarkFlagRequired("event-json")

	return cmd
}

func calendarAPIUpdateCmd() *cobra.Command {
	var calendarID, eventJSON string

	cmd := &cobra.Command{
		Use:   "update <event-id>",
		Short: "Update an event from JSON",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := calendarWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.UpdateRaw(calendarID, args[0], eventJSON)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&calendarID, "calendar", "", "Calendar ID (default: primary)")
	cmd.Flags().StringVar(&eventJSON, "event-json", "", "Event JSON (required)")

	cmd.MarkFlagRequired("event-json")

	return cmd
}

func calendarAPIPatchCmd() *cobra.Command {
	var calendarID, patchJSON string

	cmd := &cobra.Command{
		Use:   "patch <event-id>",
		Short: "Patch an event from JSON",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := calendarWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.PatchRaw(calendarID, args[0], patchJSON)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&calendarID, "calendar", "", "Calendar ID (default: primary)")
	cmd.Flags().StringVar(&patchJSON, "patch-json", "", "Patch JSON (required)")

	cmd.MarkFlagRequired("patch-json")

	return cmd
}

func calendarAPIDeleteCmd() *cobra.Command {
	var calendarID, sendUpdates string

	cmd := &cobra.Command{
		Use:   "delete <event-id>",
		Short: "Delete an event",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := calendarWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			if err := svc.Delete(calendarID, args[0], sendUpdates); err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"event_id": args[0],
				"deleted":  true,
			}, "write")
		},
	}

	cmd.Flags().StringVar(&calendarID, "calendar", "", "Calendar ID (default: primary)")
	cmd.Flags().StringVar(&sendUpdates, "send-updates", "all", "Send updates: all, externalOnly, none")

	return cmd
}

func calendarAPIQuickAddCmd() *cobra.Command {
	var calendarID, text string

	cmd := &cobra.Command{
		Use:   "quick-add",
		Short: "Quick add event from text",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := calendarWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.QuickAdd(calendarID, text)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&calendarID, "calendar", "", "Calendar ID (default: primary)")
	cmd.Flags().StringVar(&text, "text", "", "Event text (required)")

	cmd.MarkFlagRequired("text")

	return cmd
}
