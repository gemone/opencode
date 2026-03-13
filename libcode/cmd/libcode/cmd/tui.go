// Package cmd provides CLI commands for libcode
package cmd

import (
	"github.com/spf13/cobra"
)

// NewTuiCommand creates the TUI command
func NewTuiCommand(cfg interface{}) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Launch the terminal user interface",
		Long:  `Launch the interactive terminal user interface for libcode.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Launch TUI
			return nil
		},
	}

	return cmd
}
