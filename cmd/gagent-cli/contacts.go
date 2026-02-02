package main

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/ulfhaga/gagent-cli/internal/auth"
	"github.com/ulfhaga/gagent-cli/internal/config"
	"github.com/ulfhaga/gagent-cli/internal/contacts"
	"github.com/ulfhaga/gagent-cli/internal/output"
)

// contactsReadService creates a Contacts service with read scope.
func contactsReadService(ctx context.Context) (*contacts.Service, error) {
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

	return contacts.NewService(ctx, client)
}

// contactsWriteService creates a Contacts service with write scope.
func contactsWriteService(ctx context.Context) (*contacts.Service, error) {
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

	return contacts.NewService(ctx, client)
}

// contactsCmd returns the contacts command group.
func contactsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contacts",
		Short: "Google Contacts commands",
		Long:  "Read and manage Google Contacts using the People API.",
	}

	// Task commands
	cmd.AddCommand(contactsListCmd())
	cmd.AddCommand(contactsGetCmd())
	cmd.AddCommand(contactsSearchCmd())
	cmd.AddCommand(contactsGroupsCmd())
	cmd.AddCommand(contactsCreateCmd())
	cmd.AddCommand(contactsUpdateCmd())
	cmd.AddCommand(contactsDeleteCmd())

	// API commands
	cmd.AddCommand(contactsAPICmd())

	return cmd
}

func contactsListCmd() *cobra.Command {
	var limit int64

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List contacts",
		Long:  "Returns a list of contacts from Google Contacts.",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := contactsReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			contactsList, nextPageToken, err := svc.List(contacts.ListOptions{
				PageSize: limit,
			})
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"contacts":        contactsList,
				"count":           len(contactsList),
				"next_page_token": nextPageToken,
			}, "read")
		},
	}

	cmd.Flags().Int64VarP(&limit, "limit", "n", 10, "Maximum number of contacts to return")

	return cmd
}

func contactsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <resource-name>",
		Short: "Get contact details",
		Long:  "Returns full details for a specific contact.\n\nResource names are like: people/c1234567890123456789",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := contactsReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			contact, err := svc.Get(args[0])
			if err != nil {
				output.NotFoundError("Contact", args[0])
				return
			}

			output.Success(contact, "read")
		},
	}

	return cmd
}

func contactsSearchCmd() *cobra.Command {
	var limit int64

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search contacts",
		Long:  "Search contacts by name, email, or phone number.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := contactsReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			results, err := svc.Search(args[0], limit)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"query":    args[0],
				"contacts": results,
				"count":    len(results),
			}, "read")
		},
	}

	cmd.Flags().Int64VarP(&limit, "limit", "n", 10, "Maximum number of results (max 30)")

	return cmd
}

func contactsGroupsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "groups",
		Short: "List contact groups",
		Long:  "Returns a list of contact groups (labels).",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := contactsReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			groups, err := svc.Groups()
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"groups": groups,
				"count":  len(groups),
			}, "read")
		},
	}
}

func contactsCreateCmd() *cobra.Command {
	var givenName, familyName, email, phone string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a contact",
		Long:  "Creates a new contact in Google Contacts.",
		Run: func(cmd *cobra.Command, args []string) {
			if givenName == "" && familyName == "" {
				output.InvalidInputError("At least --given-name or --family-name is required")
				return
			}

			ctx := context.Background()
			svc, err := contactsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.Create(contacts.CreateContactOptions{
				GivenName:  givenName,
				FamilyName: familyName,
				Email:      email,
				Phone:      phone,
			})
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&givenName, "given-name", "", "First/given name")
	cmd.Flags().StringVar(&familyName, "family-name", "", "Last/family name")
	cmd.Flags().StringVar(&email, "email", "", "Email address")
	cmd.Flags().StringVar(&phone, "phone", "", "Phone number")

	return cmd
}

func contactsUpdateCmd() *cobra.Command {
	var givenName, familyName, email, phone string

	cmd := &cobra.Command{
		Use:   "update <resource-name>",
		Short: "Update a contact",
		Long:  "Updates an existing contact.\n\nResource names are like: people/c1234567890123456789",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			opts := contacts.UpdateContactOptions{}
			hasUpdate := false

			if cmd.Flags().Changed("given-name") {
				opts.GivenName = &givenName
				hasUpdate = true
			}
			if cmd.Flags().Changed("family-name") {
				opts.FamilyName = &familyName
				hasUpdate = true
			}
			if cmd.Flags().Changed("email") {
				opts.Email = &email
				hasUpdate = true
			}
			if cmd.Flags().Changed("phone") {
				opts.Phone = &phone
				hasUpdate = true
			}

			if !hasUpdate {
				output.InvalidInputError("At least one field to update is required")
				return
			}

			ctx := context.Background()
			svc, err := contactsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.Update(args[0], opts)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&givenName, "given-name", "", "First/given name")
	cmd.Flags().StringVar(&familyName, "family-name", "", "Last/family name")
	cmd.Flags().StringVar(&email, "email", "", "Email address")
	cmd.Flags().StringVar(&phone, "phone", "", "Phone number")

	return cmd
}

func contactsDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <resource-name>",
		Short: "Delete a contact",
		Long:  "Permanently deletes a contact.\n\nResource names are like: people/c1234567890123456789",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := contactsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			if err := svc.Delete(args[0]); err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"resource_name": args[0],
				"deleted":       true,
			}, "write")
		},
	}

	return cmd
}

// Contacts API commands
func contactsAPICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Low-level Contacts API commands",
		Long:  "Direct access to People API operations.",
	}

	cmd.AddCommand(contactsAPIGetCmd())
	cmd.AddCommand(contactsAPIListCmd())
	cmd.AddCommand(contactsAPICreateCmd())
	cmd.AddCommand(contactsAPIUpdateCmd())
	cmd.AddCommand(contactsAPIDeleteCmd())

	return cmd
}

func contactsAPIGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <resource-name>",
		Short: "Get a contact",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := contactsReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			contact, err := svc.Get(args[0])
			if err != nil {
				output.NotFoundError("Contact", args[0])
				return
			}

			output.Success(contact, "read")
		},
	}

	return cmd
}

func contactsAPIListCmd() *cobra.Command {
	var pageToken string
	var pageSize int64

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List contacts",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := contactsReadService(ctx)
			if err != nil {
				output.FailureFromError(output.ErrAuthRequired, err)
				return
			}

			contactsList, nextPageToken, err := svc.List(contacts.ListOptions{
				PageSize:  pageSize,
				PageToken: pageToken,
			})
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"contacts":        contactsList,
				"count":           len(contactsList),
				"next_page_token": nextPageToken,
			}, "read")
		},
	}

	cmd.Flags().Int64VarP(&pageSize, "page-size", "n", 10, "Page size")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "Page token for pagination")

	return cmd
}

func contactsAPICreateCmd() *cobra.Command {
	var personJSON string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a contact from JSON",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := contactsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.CreateRaw(personJSON)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&personJSON, "json", "", "Contact JSON (required)")
	cmd.MarkFlagRequired("json")

	return cmd
}

func contactsAPIUpdateCmd() *cobra.Command {
	var personJSON, updateFields string

	cmd := &cobra.Command{
		Use:   "update <resource-name>",
		Short: "Update a contact from JSON",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := contactsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			result, err := svc.UpdateRaw(args[0], personJSON, updateFields)
			if err != nil {
				output.APIError(err)
				return
			}

			output.Success(result, "write")
		},
	}

	cmd.Flags().StringVar(&personJSON, "json", "", "Contact JSON (required)")
	cmd.Flags().StringVar(&updateFields, "update-fields", "names,emailAddresses,phoneNumbers", "Fields to update (comma-separated)")
	cmd.MarkFlagRequired("json")

	return cmd
}

func contactsAPIDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <resource-name>",
		Short: "Delete a contact",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			svc, err := contactsWriteService(ctx)
			if err != nil {
				output.Failure(output.ErrScopeInsufficient, err.Error(), nil)
				return
			}

			if err := svc.Delete(args[0]); err != nil {
				output.APIError(err)
				return
			}

			output.Success(map[string]interface{}{
				"resource_name": args[0],
				"deleted":       true,
			}, "write")
		},
	}

	return cmd
}
