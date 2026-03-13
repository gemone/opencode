// Package cmd provides CLI commands for libcode
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewSessionCommand creates the session command
func NewSessionCommand(cfg interface{}) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session [command]",
		Short: "Manage sessions",
		Long:  `Manage chat sessions and their history.`,
	}

	// Add subcommands
	cmd.AddCommand(newSessionListCommand())
	cmd.AddCommand(newSessionShowCommand())
	cmd.AddCommand(newSessionDeleteCommand())

	return cmd
}

// newSessionListCommand creates the session list command
func newSessionListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Sessions:")
			// TODO: List sessions from database
			return nil
		},
	}

	return cmd
}

// newSessionShowCommand creates the session show command
func newSessionShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [session-id]",
		Short: "Show session details",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("session ID is required")
			}
			fmt.Printf("Session details for: %s\n", args[0])
			// TODO: Show session details
			return nil
		},
	}

	return cmd
}

// newSessionDeleteCommand creates the session delete command
func newSessionDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [session-id]",
		Short: "Delete a session",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("session ID is required")
			}
			fmt.Printf("Deleting session: %s\n", args[0])
			// TODO: Delete session
			return nil
		},
	}

	return cmd
}
