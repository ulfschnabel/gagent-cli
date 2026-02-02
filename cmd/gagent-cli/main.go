// Package main is the entry point for gagent-cli.
package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/ulfhaga/gagent-cli/internal/output"
)

var (
	version = "0.1.0"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "gagent-cli",
		Short: "A CLI tool for AI agents to access Google APIs",
		Long: `gagent-cli provides safe, scoped access to Google Mail, Calendar,
Docs, Sheets, and Slides APIs for AI agents.

Read and write operations require separate authorization flows.
A user can authorize read-only access without ever granting write access.`,
		Version: version,
	}

	// Disable default completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Add subcommands
	rootCmd.AddCommand(authCmd())
	rootCmd.AddCommand(gmailCmd())
	rootCmd.AddCommand(calendarCmd())
	rootCmd.AddCommand(contactsCmd())
	rootCmd.AddCommand(docsCmd())
	rootCmd.AddCommand(sheetsCmd())
	rootCmd.AddCommand(slidesCmd())
	rootCmd.AddCommand(configCmd())

	if err := rootCmd.Execute(); err != nil {
		output.FailureFromError(output.ErrInternal, err)
		os.Exit(1)
	}
}

// authCmd returns the auth command group.
func authCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication commands",
		Long:  "Manage OAuth authentication with Google APIs.",
	}

	cmd.AddCommand(authSetupCmd())
	cmd.AddCommand(authLoginCmd())
	cmd.AddCommand(authStatusCmd())
	cmd.AddCommand(authRevokeCmd())

	return cmd
}

// gmailCmd returns the gmail command group.
func gmailCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gmail",
		Short: "Gmail commands",
		Long:  "Read and manage Gmail messages.",
	}

	// Task commands
	cmd.AddCommand(gmailInboxCmd())
	cmd.AddCommand(gmailReadCmd())
	cmd.AddCommand(gmailSearchCmd())
	cmd.AddCommand(gmailThreadCmd())
	cmd.AddCommand(gmailSendCmd())
	cmd.AddCommand(gmailReplyCmd())
	cmd.AddCommand(gmailForwardCmd())
	cmd.AddCommand(gmailDraftCmd())

	// API commands
	cmd.AddCommand(gmailAPICmd())

	return cmd
}

// calendarCmd returns the calendar command group.
func calendarCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "calendar",
		Short: "Calendar commands",
		Long:  "Read and manage Google Calendar events.",
	}

	// Task commands
	cmd.AddCommand(calendarTodayCmd())
	cmd.AddCommand(calendarWeekCmd())
	cmd.AddCommand(calendarUpcomingCmd())
	cmd.AddCommand(calendarEventCmd())
	cmd.AddCommand(calendarFindCmd())
	cmd.AddCommand(calendarFreeBusyCmd())
	cmd.AddCommand(calendarScheduleCmd())
	cmd.AddCommand(calendarRescheduleCmd())
	cmd.AddCommand(calendarCancelCmd())
	cmd.AddCommand(calendarRespondCmd())

	// API commands
	cmd.AddCommand(calendarAPICmd())

	return cmd
}

// docsCmd returns the docs command group.
func docsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Google Docs commands",
		Long:  "Read and manage Google Docs documents.",
	}

	// Task commands
	cmd.AddCommand(docsListCmd())
	cmd.AddCommand(docsReadCmd())
	cmd.AddCommand(docsExportCmd())
	cmd.AddCommand(docsOutlineCmd())
	cmd.AddCommand(docsCreateCmd())
	cmd.AddCommand(docsAppendCmd())
	cmd.AddCommand(docsPrependCmd())
	cmd.AddCommand(docsReplaceTextCmd())
	cmd.AddCommand(docsUpdateSectionCmd())

	// API commands
	cmd.AddCommand(docsAPICmd())

	return cmd
}

// sheetsCmd returns the sheets command group.
func sheetsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sheets",
		Short: "Google Sheets commands",
		Long:  "Read and manage Google Sheets spreadsheets.",
	}

	// Task commands
	cmd.AddCommand(sheetsListCmd())
	cmd.AddCommand(sheetsReadCmd())
	cmd.AddCommand(sheetsInfoCmd())
	cmd.AddCommand(sheetsExportCmd())
	cmd.AddCommand(sheetsQueryCmd())
	cmd.AddCommand(sheetsCreateCmd())
	cmd.AddCommand(sheetsWriteCmd())
	cmd.AddCommand(sheetsAppendCmd())
	cmd.AddCommand(sheetsClearCmd())
	cmd.AddCommand(sheetsAddSheetCmd())
	cmd.AddCommand(sheetsDeleteSheetCmd())

	// API commands
	cmd.AddCommand(sheetsAPICmd())

	return cmd
}

// slidesCmd returns the slides command group.
func slidesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "slides",
		Short: "Google Slides commands",
		Long:  "Read and manage Google Slides presentations.",
	}

	// Task commands
	cmd.AddCommand(slidesListCmd())
	cmd.AddCommand(slidesInfoCmd())
	cmd.AddCommand(slidesReadCmd())
	cmd.AddCommand(slidesExportCmd())
	cmd.AddCommand(slidesTextCmd())
	cmd.AddCommand(slidesCreateCmd())
	cmd.AddCommand(slidesAddSlideCmd())
	cmd.AddCommand(slidesDeleteSlideCmd())
	cmd.AddCommand(slidesUpdateTextCmd())
	cmd.AddCommand(slidesAddTextCmd())
	cmd.AddCommand(slidesAddImageCmd())

	// API commands
	cmd.AddCommand(slidesAPICmd())

	return cmd
}

// configCmd returns the config command.
func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration commands",
		Long:  "Manage gagent-cli configuration.",
	}

	cmd.AddCommand(configSetCmd())
	cmd.AddCommand(configGetCmd())

	return cmd
}
