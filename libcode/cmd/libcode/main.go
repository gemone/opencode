// libcode - AI-powered development tool (Go rewrite of OpenCode)
package main

import (
	"fmt"
	"os"

	"github.com/gemone/libcode/cmd/libcode/cmd"
	"github.com/gemone/libcode/internal/config"
	"github.com/gemone/libcode/internal/logger"
	"github.com/spf13/cobra"
)

const (
	version = "0.1.0-alpha"
	name    = "libcode"
)

var (
	printLogs bool
	logLevel  string
)

func main() {
	// Initialize config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		// Continue with defaults
		cfg = &config.Config{}
	}

	// Determine log level
	logLevelValue := "INFO"
	if logLevel != "" {
		logLevelValue = logLevel
	} else if cfg.LogLevel != "" {
		logLevelValue = cfg.LogLevel
	}
	if os.Getenv("LIBCODE_DEV") == "true" {
		logLevelValue = "DEBUG"
	}

	// Initialize logger
	if err := logger.Init(logLevelValue, ""); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing logger: %v\n", err)
		os.Exit(1)
	}

	// Execute root command
	if err := rootCmd(cfg).Execute(); err != nil {
		os.Exit(1)
	}
}

// rootCmd represents the base command when called without any subcommands
func rootCmd(cfg *config.Config) *cobra.Command {
	root := &cobra.Command{
		Use:   name,
		Short: "AI-powered development tool",
		Long: `libcode is an AI-powered development tool written in Go.
It provides code intelligence, tool execution, and AI agent capabilities.`,
		Version: version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Set environment variables
			os.Setenv("AGENT", "1")
			os.Setenv("LIBCODE", "1")
		},
	}

	// Global flags
	root.PersistentFlags().BoolVar(&printLogs, "print-logs", false, "print logs to stderr")
	root.PersistentFlags().StringVar(&logLevel, "log-level", "", "log level (DEBUG, INFO, WARN, ERROR)")

	// Add subcommands
	root.AddCommand(cmd.NewRunCommand(cfg))
	root.AddCommand(cmd.NewTuiCommand(cfg))
	root.AddCommand(cmd.NewAuthCommand(cfg))
	root.AddCommand(cmd.NewAgentCommand(cfg))
	root.AddCommand(cmd.NewMcpCommand(cfg))
	root.AddCommand(cmd.NewGithubCommand(cfg))
	root.AddCommand(cmd.NewModelsCommand(cfg))
	root.AddCommand(cmd.NewSessionCommand(cfg))
	root.AddCommand(cmd.NewDbCommand(cfg))
	root.AddCommand(cmd.NewVersionCommand(cfg))
	root.AddCommand(cmd.NewCompletionCommand())

	return root
}
