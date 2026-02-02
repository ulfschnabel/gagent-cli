package main

import (
	"github.com/spf13/cobra"
	"github.com/ulfhaga/gagent-cli/internal/config"
	"github.com/ulfhaga/gagent-cli/internal/output"
)

func configSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set a configuration value",
		Long: `Set a configuration value.

Available keys:
  default_calendar  - Default calendar ID (default: "primary")
  audit_log         - Enable audit logging (true/false)`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]
			value := args[1]

			if err := config.Set(key, value); err != nil {
				output.InvalidInputError(err.Error())
				return
			}

			output.SuccessNoScope(map[string]interface{}{
				"key":   key,
				"value": value,
				"set":   true,
			})
		},
	}
}

func configGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [key]",
		Short: "Get a configuration value",
		Long: `Get a configuration value.

Available keys:
  client_id         - OAuth client ID
  client_secret     - OAuth client secret
  default_calendar  - Default calendar ID
  output_format     - Output format
  audit_log         - Audit logging enabled`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]

			value, err := config.Get(key)
			if err != nil {
				output.InvalidInputError(err.Error())
				return
			}

			output.SuccessNoScope(map[string]interface{}{
				"key":   key,
				"value": value,
			})
		},
	}
}
