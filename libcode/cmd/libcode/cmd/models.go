// Package cmd provides CLI commands for libcode
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewModelsCommand creates the models command
func NewModelsCommand(cfg interface{}) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "models [command]",
		Short: "Manage AI models",
		Long:  `Manage AI model configurations and availability.`,
	}

	// Add subcommands
	cmd.AddCommand(newModelsListCommand())
	cmd.AddCommand(newModelsInfoCommand())

	return cmd
}

// newModelsListCommand creates the models list command
func newModelsListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available models",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Available models:")
			fmt.Println("  - claude-3-5-sonnet-20241022: Claude 3.5 Sonnet")
			fmt.Println("  - claude-3-5-haiku-20241022: Claude 3.5 Haiku")
			fmt.Println("  - gpt-4-turbo: GPT-4 Turbo")
			fmt.Println("  - gpt-3.5-turbo: GPT-3.5 Turbo")
			// TODO: List from configuration
			return nil
		},
	}

	return cmd
}

// newModelsInfoCommand creates the models info command
func newModelsInfoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info [model]",
		Short: "Show information about a model",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("model name is required")
			}
			fmt.Printf("Model info for: %s\n", args[0])
			// TODO: Show model info
			return nil
		},
	}

	return cmd
}
