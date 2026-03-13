// Package tui provides TUI execution for libcode
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the TUI application
func Run() error {
	model := NewModel()
	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := program.Run(); err != nil {
		return err
	}

	return nil
}

// RunWithOptions starts the TUI with custom options
func RunWithOptions(opts ...tea.ProgramOption) error {
	model := NewModel()
	options := []tea.ProgramOption{
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	}
	options = append(options, opts...)

	program := tea.NewProgram(model, options...)

	if _, err := program.Run(); err != nil {
		return err
	}

	return nil
}
