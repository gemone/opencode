// Package builtin provides built-in tool implementations
package builtin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gemone/libcode/internal/tool/framework"
)

const (
	defaultBashTimeout = 2 * time.Minute
	maxMetadataLength  = 30000
)

// BashTool executes shell commands
type BashTool struct {
	workingDir string
}

// NewBashTool creates a new bash tool
func NewBashTool(workingDir string) *BashTool {
	return &BashTool{workingDir: workingDir}
}

// ID returns the tool identifier
func (t *BashTool) ID() string {
	return "bash"
}

// Description returns the tool description
func (t *BashTool) Description() string {
	return "Execute shell commands with timeout control. Supports command execution in the working directory with permission checks."
}

// Parameters returns the parameter schema
func (t *BashTool) Parameters() *framework.Schema {
	return framework.NewSchema().
		AddProperty("command", framework.Property{
			Type:        "string",
			Description: "The command to execute",
			Required:    true,
		}).
		AddProperty("description", framework.Property{
			Type:        "string",
			Description: "Clear, concise description of what this command does in 5-10 words. Examples: Input: 'ls' Output: 'Lists files in current directory'",
			Required:    true,
		}).
		AddProperty("timeout", framework.Property{
			Type:        "integer",
			Description: "Optional timeout in milliseconds",
			Default:     int(defaultBashTimeout.Milliseconds()),
		}).
		AddProperty("workdir", framework.Property{
			Type:        "string",
			Description: fmt.Sprintf("The working directory to run the command in. Defaults to %s", t.workingDir),
			Format:      "uri",
		})
}

// Execute executes the bash command
func (t *BashTool) Execute(ctx context.Context, params map[string]any, execCtx *framework.ExecutionContext) (*framework.Result, error) {
	command, _ := params["command"].(string)
	description, _ := params["description"].(string)
	timeoutMs := int(params["timeout"].(float64))
	workdir, _ := params["workdir"].(string)

	if timeoutMs <= 0 {
		timeoutMs = int(defaultBashTimeout.Milliseconds())
	}

	if workdir == "" {
		workdir = t.workingDir
	}

	// Resolve working directory
	workdir = resolvePath(workdir, t.workingDir)

	// Check for dangerous patterns
	if err := t.checkCommandSafety(command); err != nil {
		return nil, err
	}

	// Ask for permissions based on command analysis
	if err := t.requestPermissions(ctx, command, workdir, execCtx); err != nil {
		return nil, err
	}

	// Create command with shell
	shell := "/bin/bash"
	if _, err := exec.LookPath("bash"); err != nil {
		shell = "/bin/sh" // Fallback to sh
	}

	cmd := exec.CommandContext(ctx, shell, "-c", command)
	cmd.Dir = workdir

	// Set up environment
	cmd.Env = append(os.Environ(),
		"PATH="+os.Getenv("PATH"),
		"HOME="+os.Getenv("HOME"),
	)

	// Start command
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	// Read output with timeout
	outputChan := make(chan string, 1)
	errorChan := make(chan string, 1)
	doneChan := make(chan error, 2)

	// Read stdout
	go func() {
		var output strings.Builder
		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				output.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}
		outputChan <- output.String()
	}()

	// Read stderr
	go func() {
		var errStr strings.Builder
		buf := make([]byte, 4096)
		for {
			n, err := stderr.Read(buf)
			if n > 0 {
				errStr.Write(buf[:n])
			}
			if err != nil {
				break
			}
		}
		errorChan <- errStr.String()
	}()

	// Wait for completion with timeout
	go func() {
		err := cmd.Wait()
		doneChan <- err
	}()

	// Set up timeout
	timer := time.NewTimer(time.Duration(timeoutMs) * time.Millisecond)
	defer timer.Stop()

	var output string
	var errors string
	var exitCode int
	var timedOut bool
	var userAborted bool

	select {
	case <-timer.C:
		// Timeout - kill process
		t.killProcessGroup(cmd)
		timedOut = true
		<-doneChan // Wait for Wait() to complete
		<-outputChan
		<-errorChan

	case <-ctx.Done():
		// Context cancelled
		t.killProcessGroup(cmd)
		userAborted = true
		<-doneChan
		<-outputChan
		<-errorChan

	case err := <-doneChan:
		timer.Stop()
		output = <-outputChan
		errors = <-errorChan
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode = exitError.ExitCode()
			}
		}
	}

	// Combine output
	fullOutput := output
	if errors != "" {
		fullOutput += "\n" + errors
	}

	// Add metadata for timeout or abort
	metadata := map[string]any{
		"exit":       exitCode,
		"description": description,
		"output":     truncateOutput(output, maxMetadataLength),
	}

	if timedOut {
		fullOutput += fmt.Sprintf("\n\n<bash_metadata>\nCommand terminated after exceeding timeout %d ms", timeoutMs)
		metadata["timedOut"] = true
	} else if userAborted {
		fullOutput += "\n\n<bash_metadata>\nUser aborted the command"
		metadata["aborted"] = true
	}

	return &framework.Result{
		Title:  description,
		Output: fullOutput,
		Metadata: metadata,
	}, nil
}

// checkCommandSafety checks if the command contains dangerous patterns
func (t *BashTool) checkCommandSafety(command string) error {
	// Check for commands that could modify critical system files
	dangerousPatterns := []string{
		"rm -rf /",
		"rm -rf /*",
		":(){ :|:& };:",
		"chmod 000 /",
	}

	commandLower := strings.ToLower(command)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(commandLower, pattern) {
			return fmt.Errorf("command contains dangerous pattern: %s", pattern)
		}
	}

	return nil
}

// requestPermissions analyzes the command and requests appropriate permissions
func (t *BashTool) requestPermissions(ctx context.Context, command, workdir string, execCtx *framework.ExecutionContext) error {
	if execCtx.PermissionAsker == nil {
		return nil
	}

	// Analyze command to determine what permissions are needed
	needsBash := true
	needsExternalDir := !t.containsPath(workdir)

	// Parse command to find files/directories being accessed
	tokens := strings.Fields(command)
	if len(tokens) > 0 {
		// Check for file operations
		switch tokens[0] {
		case "rm", "cp", "mv", "mkdir", "touch", "chmod", "chown", "cat", "head", "tail":
			needsExternalDir = true
		case "cd":
			// cd is handled by workdir parameter
			needsBash = false
		}
	}

	// Request bash permission if needed
	if needsBash {
		err := execCtx.PermissionAsker.Ask(&framework.PermissionRequest{
			Permission: "bash",
			Patterns:   []string{command},
			Always:     t.buildAlwaysPatterns(command),
			Metadata: map[string]any{
				"command": command,
				"workdir": workdir,
			},
		})
		if err != nil {
			return err
		}
	}

	// Request external directory permission if needed
	if needsExternalDir {
		globPattern := filepath.Join(workdir, "*")
		err := execCtx.PermissionAsker.Ask(&framework.PermissionRequest{
			Permission: "external_directory",
			Patterns:   []string{globPattern},
			Always:     []string{globPattern},
			Metadata: map[string]any{
				"workdir": workdir,
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// buildAlwaysPatterns builds "always" permission patterns for bash commands
func (t *BashTool) buildAlwaysPatterns(command string) []string {
	tokens := strings.Fields(command)
	if len(tokens) == 0 {
		return []string{"*"}
	}

	// Build prefix patterns
	var always []string
	for i := range tokens {
		if i > 0 {
			always = append(always, strings.Join(tokens[:i], " ")+" *")
		} else {
			always = append(always, tokens[0]+" *")
		}
	}

	return always
}

// killProcessGroup kills the process group for the command
func (t *BashTool) killProcessGroup(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}

	// Get process group ID
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err != nil {
		// Try killing just the process
		cmd.Process.Kill()
		return
	}

	// Kill the process group
	syscall.Kill(-pgid, syscall.SIGTERM)
}

// containsPath checks if a path is within the working directory
func (t *BashTool) containsPath(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	absWorkingDir, err := filepath.Abs(t.workingDir)
	if err != nil {
		return false
	}

	relPath, err := filepath.Rel(absWorkingDir, absPath)
	if err != nil {
		return false
	}

	return !strings.HasPrefix(relPath, "..")
}

// truncateOutput truncates output to a maximum length
func truncateOutput(output string, maxLen int) string {
	if len(output) <= maxLen {
		return output
	}
	return output[:maxLen] + "\n\n..."
}
