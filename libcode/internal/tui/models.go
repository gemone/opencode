// Package tui provides TUI components for libcode
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// chatModel represents the chat interface
type chatModel struct {
	messages []Message
	width    int
	height   int
	scroll   int
}

// Message represents a chat message
type Message struct {
	Role    string
	Content string
	Time    string
}

// newChatModel creates a new chat model
func newChatModel() chatModel {
	return chatModel{
		messages: []Message{},
		scroll:   0,
	}
}

// Init initializes the chat model
func (m chatModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the chat model
func (m chatModel) Update(msg tea.Msg) (chatModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case UserInputMsg:
		// Add user message
		m.messages = append(m.messages, Message{
			Role:    "user",
			Content: msg.Text,
			Time:    msg.Time,
		})
		// TODO: Trigger AI response
	case AssistantResponseMsg:
		// Add assistant message
		m.messages = append(m.messages, Message{
			Role:    "assistant",
			Content: msg.Text,
			Time:    msg.Time,
		})
	}
	return m, nil
}

// View renders the chat interface
func (m chatModel) View() string {
	if len(m.messages) == 0 {
		return welcomeMessage(m.width, m.height)
	}

	var result string
	for _, msg := range m.messages {
		result += renderMessage(msg, m.width)
	}
	return result
}

// renderMessage renders a single message
func renderMessage(msg Message, width int) string {
	roleStyle := lipgloss.NewStyle()
	if msg.Role == "user" {
		roleStyle = roleStyle.Foreground(lipgloss.Color("86")) // green
	} else {
		roleStyle = roleStyle.Foreground(lipgloss.Color("147")) // purple
	}

	role := roleStyle.Render("▌ " + msg.Role)
	content := lipgloss.NewStyle().Width(width - 4).Render(msg.Content)

	return role + "\n" + content + "\n\n"
}

// welcomeMessage renders the welcome message
func welcomeMessage(width, height int) string {
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("147")).
		Bold(true).
		Render("libcode")

	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("AI-powered development tool")

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("242")).
		Render("Press Ctrl+C to quit, Ctrl+O to open options")

	return "\n\n" + center(title, width) + "\n" +
		center(subtitle, width) + "\n\n\n" +
		center(help, width)
}

func center(text string, width int) string {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(text)
}

// statusBarModel represents the status bar
type statusBarModel struct {
	width      int
	status     string
	working    bool
}

// newStatusBarModel creates a new status bar model
func newStatusBarModel() statusBarModel {
	return statusBarModel{
		status:  "Ready",
		working: false,
	}
}

// Init initializes the status bar model
func (m statusBarModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the status bar
func (m statusBarModel) Update(msg tea.Msg) (statusBarModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case StatusUpdateMsg:
		m.status = msg.Status
		m.working = msg.Working
	}
	return m, nil
}

// View renders the status bar
func (m statusBarModel) View() string {
	left := m.status
	if m.working {
		left = "⟳ " + left
	}

	right := "libcode v0.1.0-alpha"

	leftStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("242"))

	rightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	return lipgloss.JoinHorizontal(lipgloss.Top,
		leftStyle.Render(left),
		lipgloss.NewStyle().Width(m.width - len(left) - len(right)).Render(""),
		rightStyle.Render(right),
	)
}

// inputModel represents the input area
type inputModel struct {
	width   int
	text    string
	cursor  int
	focused bool
}

// newInputModel creates a new input model
func newInputModel() inputModel {
	return inputModel{
		text:    "",
		cursor:  0,
		focused: true,
	}
}

// Init initializes the input model
func (m inputModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the input model
func (m inputModel) Update(msg tea.Msg) (inputModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}

		switch msg.String() {
		case "enter":
			if m.text != "" {
				// Emit input message
				return m, func() tea.Msg {
					return UserInputMsg{
						Text: m.text,
						Time: currentTime(),
					}
				}
			}
		case "backspace":
			if m.cursor > 0 {
				m.text = m.text[:m.cursor-1] + m.text[m.cursor:]
				m.cursor--
			}
		case "left":
			if m.cursor > 0 {
				m.cursor--
			}
		case "right":
			if m.cursor < len(m.text) {
				m.cursor++
			}
		default:
			// Regular character input
			if len(msg.String()) == 1 {
				m.text = m.text[:m.cursor] + msg.String() + m.text[m.cursor:]
				m.cursor++
			}
		}
	}
	return m, nil
}

// View renders the input area
func (m inputModel) View() string {
	prompt := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Render("⟩ ")

	input := m.text
	if m.cursor < len(m.text) {
		input = m.text[:m.cursor] + "▏" + m.text[m.cursor:]
	} else {
		input = m.text + "▏"
	}

	return prompt + lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Render(input)
}

// Messages

// UserInputMsg is emitted when user submits input
type UserInputMsg struct {
	Text string
	Time string
}

// AssistantResponseMsg is emitted when assistant responds
type AssistantResponseMsg struct {
	Text string
	Time string
}

// StatusUpdateMsg updates the status bar
type StatusUpdateMsg struct {
	Status  string
	Working bool
}

// currentTime returns current time as string
func currentTime() string {
	return "now" // TODO: implement proper time formatting
}
