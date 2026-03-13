// Package tui provides the terminal user interface for libcode
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the TUI application state
type Model struct {
	width      int
	height     int
	quitting   bool
	currentView ViewType

	// View models
	chat       chatModel
	statusBar  statusBarModel
	input      inputModel
}

// ViewType represents different views in the TUI
type ViewType int

const (
	ViewChat ViewType = iota
	ViewFileBrowser
	ViewSettings
)

// NewModel creates a new TUI model
func NewModel() Model {
	return Model{
		currentView: ViewChat,
		chat:        newChatModel(),
		statusBar:   newStatusBarModel(),
		input:       newInputModel(),
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		m.chat.Init(),
		m.statusBar.Init(),
		m.input.Init(),
	)
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "ctrl+b":
			m.currentView = ViewFileBrowser
		case "ctrl+s":
			m.currentView = ViewSettings
		case "esc":
			m.currentView = ViewChat
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update child models
		var cmd tea.Cmd
		m.chat, cmd = m.chat.Update(msg)
		cmds = append(cmds, cmd)

		m.statusBar, cmd = m.statusBar.Update(msg)
		cmds = append(cmds, cmd)

		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update based on current view
	switch m.currentView {
	case ViewChat:
		var cmd tea.Cmd
		m.chat, cmd = m.chat.Update(msg)
		cmds = append(cmds, cmd)

		// Pass input to input model
		if shouldHandleInput(msg) {
			m.input, cmd = m.input.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	// Always update status bar
	statusBar, cmd := m.statusBar.Update(msg)
	m.statusBar = statusBar
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the TUI
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	// Calculate layout
	contentHeight := m.height - 3 // minus status bar and input

	// Render based on current view
	var content string
	switch m.currentView {
	case ViewChat:
		content = m.chat.View()
	case ViewFileBrowser:
		content = fileBrowserView(m.width, contentHeight)
	case ViewSettings:
		content = settingsView(m.width, contentHeight)
	}

	// Combine views
	return layoutView(
		content,
		m.statusBar.View(),
		m.input.View(),
		m.width,
		m.height,
	)
}

// shouldHandleInput checks if a message should be handled by the input model
func shouldHandleInput(msg tea.Msg) bool {
	switch msg.(type) {
	case tea.KeyMsg:
		return true
	default:
		return false
	}
}

// layoutView combines the different view sections
func layoutView(content, statusBar, input string, width, height int) string {
	// This is a simplified layout - in a real implementation,
	// you'd use lipgloss for proper styling and positioning
	result := content + "\n" + statusBar + "\n" + input
	return result
}

// fileBrowserView renders the file browser
func fileBrowserView(width, height int) string {
	return "File Browser - Coming Soon\n\nPress Ctrl+B to toggle browser, Esc to return to chat"
}

// settingsView renders the settings view
func settingsView(width, height int) string {
	return "Settings - Coming Soon\n\nPress Ctrl+S to toggle settings, Esc to return to chat"
}
