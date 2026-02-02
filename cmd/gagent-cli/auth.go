package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ulfhaga/gagent-cli/internal/auth"
	"github.com/ulfhaga/gagent-cli/internal/config"
	"github.com/ulfhaga/gagent-cli/internal/output"
	"golang.org/x/oauth2"
)

func authSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Interactive setup wizard",
		Long:  "Guide through Google Cloud project creation and OAuth credential setup.",
		Run: func(cmd *cobra.Command, args []string) {
			runSetupWizard()
		},
	}
}

func runSetupWizard() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("Welcome to gagent-cli setup!")
	fmt.Println()
	fmt.Println("This tool requires OAuth credentials from a Google Cloud project.")
	fmt.Println("I'll guide you through the setup process.")
	fmt.Println()

	// Step 1: Create Google Cloud Project
	fmt.Println("Step 1: Create a Google Cloud Project")
	fmt.Println("--------------------------------------")
	fmt.Println("Open: https://console.cloud.google.com/projectcreate")
	fmt.Println()
	fmt.Println("1. Enter a project name (e.g., \"gagent-cli\")")
	fmt.Println("2. Click \"Create\"")
	fmt.Println()
	fmt.Print("Press Enter when done...")
	reader.ReadString('\n')
	fmt.Println()

	// Step 2: Enable APIs
	fmt.Println("Step 2: Enable APIs")
	fmt.Println("-------------------")
	fmt.Println("Enable each API (click \"Enable\" on each page):")
	fmt.Println()
	fmt.Println("  https://console.cloud.google.com/apis/library/gmail.googleapis.com")
	fmt.Println("  https://console.cloud.google.com/apis/library/calendar-json.googleapis.com")
	fmt.Println("  https://console.cloud.google.com/apis/library/people.googleapis.com")
	fmt.Println("  https://console.cloud.google.com/apis/library/docs.googleapis.com")
	fmt.Println("  https://console.cloud.google.com/apis/library/sheets.googleapis.com")
	fmt.Println("  https://console.cloud.google.com/apis/library/slides.googleapis.com")
	fmt.Println("  https://console.cloud.google.com/apis/library/drive.googleapis.com")
	fmt.Println()
	fmt.Print("Press Enter when all APIs are enabled...")
	reader.ReadString('\n')
	fmt.Println()

	// Step 3: Configure OAuth Consent Screen
	fmt.Println("Step 3: Configure OAuth Consent Screen")
	fmt.Println("--------------------------------------")
	fmt.Println("Open: https://console.cloud.google.com/apis/credentials/consent")
	fmt.Println()
	fmt.Println("1. Select \"External\" user type (or \"Internal\" if using Google Workspace)")
	fmt.Println("2. Fill in the required fields:")
	fmt.Println("   - App name: gagent-cli")
	fmt.Println("   - User support email: your email")
	fmt.Println("   - Developer contact: your email")
	fmt.Println("3. Click \"Save and Continue\"")
	fmt.Println("4. Skip the Scopes section (click \"Save and Continue\")")
	fmt.Println("5. Skip Test Users for now (click \"Save and Continue\")")
	fmt.Println("6. Review and click \"Back to Dashboard\"")
	fmt.Println()
	fmt.Print("Press Enter when done...")
	reader.ReadString('\n')
	fmt.Println()

	// Step 4: Add Test Users
	fmt.Println("Step 4: Add Yourself as a Test User")
	fmt.Println("------------------------------------")
	fmt.Println("Open: https://console.cloud.google.com/apis/credentials/consent")
	fmt.Println()
	fmt.Println("1. Scroll down to \"Test users\" section")
	fmt.Println("2. Click \"+ ADD USERS\"")
	fmt.Println("3. Enter your Google email address")
	fmt.Println("4. Click \"Save\"")
	fmt.Println()
	fmt.Println("⚠️  This step is required! Without it, you'll get an error during login.")
	fmt.Println()
	fmt.Print("Press Enter when done...")
	reader.ReadString('\n')
	fmt.Println()

	// Step 5: Create OAuth Credentials
	fmt.Println("Step 5: Create OAuth Credentials")
	fmt.Println("--------------------------------")
	fmt.Println("Open: https://console.cloud.google.com/apis/credentials")
	fmt.Println()
	fmt.Println("1. Click \"Create Credentials\" → \"OAuth client ID\"")
	fmt.Println("2. Select \"Desktop app\" as application type")
	fmt.Println("3. Name it \"gagent-cli\"")
	fmt.Println("4. Click \"Download JSON\" to save credentials (recommended for backup)")
	fmt.Println("5. Copy the Client ID and Client Secret shown on screen")
	fmt.Println()

	// Get Client ID
	fmt.Print("Enter Client ID: ")
	clientID, _ := reader.ReadString('\n')
	clientID = strings.TrimSpace(clientID)

	if clientID == "" {
		output.InvalidInputError("Client ID is required")
		return
	}

	// Get Client Secret
	fmt.Print("Enter Client Secret: ")
	clientSecret, _ := reader.ReadString('\n')
	clientSecret = strings.TrimSpace(clientSecret)

	if clientSecret == "" {
		output.InvalidInputError("Client Secret is required")
		return
	}

	// Save configuration
	cfg := config.DefaultConfig()
	cfg.ClientID = clientID
	cfg.ClientSecret = clientSecret

	if err := config.Save(cfg); err != nil {
		output.FailureFromError(output.ErrInternal, err)
		return
	}

	configDir, _ := config.GetConfigDir()
	fmt.Println()
	fmt.Printf("✓ Configuration saved to %s/config.json\n", configDir)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  gagent-cli auth login --scope read   # Authorize read access")
	fmt.Println("  gagent-cli auth login --scope write  # Authorize write access (optional)")
	fmt.Println()

	output.SuccessNoScope(map[string]string{
		"status":     "configured",
		"config_dir": configDir,
	})
}

func authLoginCmd() *cobra.Command {
	var scope string
	var manual bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authorize access",
		Long: `Run OAuth flow to authorize read or write access.

Use --manual flag when running on a remote machine or headless environment
where the browser callback cannot reach the CLI.`,
		Run: func(cmd *cobra.Command, args []string) {
			if !auth.IsValidScopeType(scope) {
				output.InvalidInputError("Invalid scope. Use 'read' or 'write'.")
				return
			}

			cfg, err := config.Load()
			if err != nil {
				output.Failure(output.ErrAuthRequired, "Configuration not found. Run: gagent-cli auth setup", nil)
				return
			}

			scopeType := auth.ScopeType(scope)
			oauthConfig := auth.NewOAuthConfig(cfg.ClientID, cfg.ClientSecret, scopeType)

			ctx := context.Background()
			var token *oauth2.Token

			if manual {
				token, err = auth.PerformOAuthFlowManual(ctx, oauthConfig)
			} else {
				token, err = auth.PerformOAuthFlow(ctx, oauthConfig)
			}

			if err != nil {
				output.FailureFromError(output.ErrInternal, err)
				return
			}

			configDir, err := config.GetConfigDir()
			if err != nil {
				output.FailureFromError(output.ErrInternal, err)
				return
			}

			tokenPath := auth.TokenPath(configDir, scopeType)
			if err := auth.SaveToken(tokenPath, token); err != nil {
				output.FailureFromError(output.ErrInternal, err)
				return
			}

			fmt.Printf("\n✓ %s access authorized successfully!\n", scope)
			output.SuccessNoScope(map[string]interface{}{
				"scope":      scope,
				"authorized": true,
				"token_path": tokenPath,
			})
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "", "Scope to authorize: 'read' or 'write' (required)")
	cmd.Flags().BoolVar(&manual, "manual", false, "Use manual flow for remote/headless environments")
	cmd.MarkFlagRequired("scope")

	return cmd
}

func authStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current auth status",
		Long:  "Display the current authentication status for read and write scopes.",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Load()
			if err != nil {
				output.Failure(output.ErrAuthRequired, "Configuration not found. Run: gagent-cli auth setup", nil)
				return
			}

			configDir, err := config.GetConfigDir()
			if err != nil {
				output.FailureFromError(output.ErrInternal, err)
				return
			}

			readAuthorized := auth.TokenExists(configDir, auth.ScopeRead)
			writeAuthorized := auth.TokenExists(configDir, auth.ScopeWrite)

			status := map[string]interface{}{
				"configured":       true,
				"config_dir":       configDir,
				"read_authorized":  readAuthorized,
				"write_authorized": writeAuthorized,
				"client_id":        maskClientID(cfg.ClientID),
			}

			output.SuccessNoScope(status)
		},
	}
}

func authRevokeCmd() *cobra.Command {
	var scope string

	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke access token",
		Long:  "Delete the stored OAuth token for the specified scope.",
		Run: func(cmd *cobra.Command, args []string) {
			if !auth.IsValidScopeType(scope) {
				output.InvalidInputError("Invalid scope. Use 'read' or 'write'.")
				return
			}

			configDir, err := config.GetConfigDir()
			if err != nil {
				output.FailureFromError(output.ErrInternal, err)
				return
			}

			scopeType := auth.ScopeType(scope)
			if err := auth.DeleteToken(configDir, scopeType); err != nil {
				output.FailureFromError(output.ErrInternal, err)
				return
			}

			output.SuccessNoScope(map[string]interface{}{
				"scope":   scope,
				"revoked": true,
			})
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "", "Scope to revoke: 'read' or 'write' (required)")
	cmd.MarkFlagRequired("scope")

	return cmd
}

// maskClientID masks part of the client ID for display.
func maskClientID(clientID string) string {
	if len(clientID) < 20 {
		return "***"
	}
	return clientID[:10] + "..." + clientID[len(clientID)-10:]
}
