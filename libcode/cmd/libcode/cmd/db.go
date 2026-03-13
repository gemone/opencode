// Package cmd provides CLI commands for libcode
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewDbCommand creates the db command
func NewDbCommand(cfg interface{}) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db [command]",
		Short: "Database operations",
		Long:  `Manage the libcode database.`,
	}

	// Add subcommands
	cmd.AddCommand(newDbStatsCommand())
	cmd.AddCommand(newDbMigrateCommand())
	cmd.AddCommand(newDbResetCommand())

	return cmd
}

// newDbStatsCommand creates the db stats command
func newDbStatsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show database statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Database statistics:")
			// TODO: Show database stats
			return nil
		},
	}

	return cmd
}

// newDbMigrateCommand creates the db migrate command
func newDbMigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Running database migrations...")
			// TODO: Run migrations
			return nil
		},
	}

	return cmd
}

// newDbResetCommand creates the db reset command
func newDbResetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset the database",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Resetting database...")
			// TODO: Reset database
			return nil
		},
	}

	return cmd
}
