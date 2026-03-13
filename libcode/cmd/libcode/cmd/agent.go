// Package cmd provides CLI commands for libcode
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewAgentCommand creates the agent command
func NewAgentCommand(cfg interface{}) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent [command]",
		Short: "Manage AI agents",
		Long:  `Manage AI agents and their configurations.`,
	}

	// Add subcommands
	cmd.AddCommand(newAgentListCommand())
	cmd.AddCommand(newAgentInfoCommand())

	return cmd
}

// newAgentListCommand creates the agent list command
func newAgentListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Available agents:")
			fmt.Println("  - claude: Anthropic Claude agent")
			fmt.Println("  - gpt: OpenAI GPT agent")
			// TODO: List from configuration
			return nil
		},
	}

	return cmd
}

// newAgentInfoCommand creates the agent info command
func newAgentInfoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info [agent]",
		Short: "Show information about an agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("agent name is required")
			}
			fmt.Printf("Agent info for: %s\n", args[0])
			// TODO: Show agent info
			return nil
		},
	}

	return cmd
}
