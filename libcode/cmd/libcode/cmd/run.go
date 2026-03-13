// Package cmd provides CLI commands for libcode
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewRunCommand creates the run command
func NewRunCommand(cfg interface{}) *cobra.Command {
	var (
		agent     string
		model     string
		nonInteractive bool
	)

	cmd := &cobra.Command{
		Use:   "run [task...]",
		Short: "Run an AI agent to complete a task",
		Long: `Run an AI agent to complete a task. The agent will use available tools
to accomplish the specified task.

If no task is provided, starts in interactive mode.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && !nonInteractive {
				// TODO: Start interactive mode
				fmt.Println("Interactive mode - coming soon")
				return nil
			}

			// TODO: Run agent with task
			fmt.Printf("Running agent: %s\n", agent)
			if model != "" {
				fmt.Printf("Using model: %s\n", model)
			}
			fmt.Printf("Task: %v\n", args)

			return nil
		},
	}

	cmd.Flags().StringVarP(&agent, "agent", "a", "", "Agent to use")
	cmd.Flags().StringVarP(&model, "model", "m", "", "Model to use")
	cmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "Disable interactive mode")

	return cmd
}
