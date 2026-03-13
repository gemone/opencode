// Package cmd provides CLI commands for libcode
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// NewVersionCommand creates the version command
func NewVersionCommand(cfg interface{}) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("libcode version 0.1.0-alpha")
		},
	}
}

// NewCompletionCommand creates the completion command
func NewCompletionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion script",
		Long: `To load completions:

Bash:
  $ source <(libcode completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ libcode completion bash > /etc/bash_completion.d/libcode
  # macOS:
  $ libcode completion bash > /usr/local/etc/bash_completion.d/libcode

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ libcode completion zsh > "${fpath[1]}/_libcode"

  # You will need to start a new shell for this setup to take effect.

fish:
  $ libcode completion fish | source

  # To load completions for each session, execute once:
  $ libcode completion fish > ~/.config/fish/completions/libcode.fish

PowerShell:
  PS> libcode completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> libcode completion powershell > libcode.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}

			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}

	return cmd
}
