package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConnectionModel represents the connection details screen
type ConnectionModel struct {
	focused        int
	inputs         []textinput.Model
	cursorMode     cursor.Mode
	width          int
	height         int
	spinner        spinner.Model
	testingConn    bool
	testComplete   bool
	testSuccessful bool
	errorMsg       string
}

// NewConnectionModel creates a new connection model
func NewConnectionModel() ConnectionModel {
	m := ConnectionModel{
		inputs: make([]textinput.Model, 3),
	}

	// URL input
	m.inputs[0] = textinput.New()
	m.inputs[0].Placeholder = "https://elasticsearch-host:9200"
	m.inputs[0].Focus()
	m.inputs[0].Width = 40
	m.inputs[0].Prompt = "URL: "

	// Username input
	m.inputs[1] = textinput.New()
	m.inputs[1].Placeholder = "elastic"
	m.inputs[1].Width = 40
	m.inputs[1].Prompt = "Username: "

	// Password input
	m.inputs[2] = textinput.New()
	m.inputs[2].Placeholder = "changeme"
	m.inputs[2].Width = 40
	m.inputs[2].Prompt = "Password: "
	m.inputs[2].EchoMode = textinput.EchoPassword
	m.inputs[2].EchoCharacter = '•'

	// Initialize cursor mode
	m.cursorMode = cursor.CursorBlink

	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	m.spinner = s

	return m
}

// SetSize updates the size of the connection model
func (m *ConnectionModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Init initializes the connection model
func (m ConnectionModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages and updates the model
func (m ConnectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "tab", "shift+tab", "up", "down":
			// Only allow switching inputs if not testing connection
			if !m.testingConn {
				s := msg.String()

				// Did the user press tab or shift+tab?
				if s == "tab" || s == "down" {
					m.focused = (m.focused + 1) % len(m.inputs)
				} else if s == "shift+tab" || s == "up" {
					m.focused = (m.focused - 1 + len(m.inputs)) % len(m.inputs)
				}

				for i := range m.inputs {
					if i == m.focused {
						cmds = append(cmds, m.inputs[i].Focus())
					} else {
						m.inputs[i].Blur()
					}
				}
			}

		case "enter":
			// Check if all fields are filled
			if !m.testingConn && !m.testComplete {
				allFilled := true
				for i := range m.inputs {
					if strings.TrimSpace(m.inputs[i].Value()) == "" {
						allFilled = false
						break
					}
				}

				if allFilled {
					m.testingConn = true
					return m, tea.Batch(
						m.spinner.Tick,
						testConnection(
							m.inputs[0].Value(),
							m.inputs[1].Value(),
							m.inputs[2].Value(),
						),
					)
				}
			} else if m.testComplete && m.testSuccessful {
				// Connection was successful, proceed
				return m, nil
			}
		}

	case spinner.TickMsg:
		if m.testingConn {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case connectionTestResultMsg:
		m.testingConn = false
		m.testComplete = true
		m.testSuccessful = msg.success
		if !msg.success {
			m.errorMsg = msg.errorMsg
		}
		return m, nil
	}

	// Only update the currently focused input
	if !m.testingConn {
		cmd := m.updateInputs(msg)
		return m, cmd
	}

	return m, tea.Batch(cmds...)
}

// updateInputs updates the text inputs
func (m *ConnectionModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	// Update the focused input
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)

	return cmd
}

// Custom test connection message
type connectionTestResultMsg struct {
	success  bool
	errorMsg string
}

// testConnection simulates testing the connection
func testConnection(url, username, password string) tea.Cmd {
	return func() tea.Msg {
		// In a real app, you would test the connection here
		// For this example, we'll just simulate a delay and success
		time.Sleep(2 * time.Second)

		// Simple validation
		if !strings.HasPrefix(url, "http") {
			return connectionTestResultMsg{
				success:  false,
				errorMsg: "URL must start with http:// or https://",
			}
		}

		return connectionTestResultMsg{
			success: true,
		}
	}
}

// View renders the connection screen
func (m ConnectionModel) View() string {
	if m.width == 0 {
		return "Loading connection screen..."
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width - 2).
		Align(lipgloss.Center)

	title := titleStyle.Render("Connection Details")

	// Build form
	var content string

	if m.testingConn {
		// Show testing message
		spinnerView := m.spinner.View() + " Testing connection..."
		spinnerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Padding(2, 0).
			Width(m.width - 2).
			Align(lipgloss.Center)

		content = spinnerStyle.Render(spinnerView)
	} else if m.testComplete {
		if m.testSuccessful {
			// Show success message
			successStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#73F273")).
				Padding(2, 0).
				Width(m.width - 2).
				Align(lipgloss.Center)

			content = successStyle.Render("✓ Connection successful! Press Enter to continue.")
		} else {
			// Show error message
			inputs := ""
			for i := range m.inputs {
				inputs += m.inputs[i].View() + "\n"
			}

			errorStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF5555")).
				Padding(1, 0)

			errorView := errorStyle.Render("✗ " + m.errorMsg)

			content = inputs + "\n" + errorView + "\n\nPlease correct the issue and try again."
		}
	} else {
		// Show inputs
		inputs := ""
		for i := range m.inputs {
			inputs += m.inputs[i].View() + "\n"
		}

		content = inputs
	}

	helpText := "• Tab to navigate • Enter to test connection •"
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Align(lipgloss.Center).
		Width(m.width - 2)

	help := helpStyle.Render(helpText)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"\n",
		content,
		"\n",
		help,
	)
}

// IsComplete returns whether connection setup is complete
func (m ConnectionModel) IsComplete() bool {
	return m.testComplete && m.testSuccessful
}

// GetConnectionDetails returns the connection details
func (m ConnectionModel) GetConnectionDetails() ConnectionDetails {
	return ConnectionDetails{
		URL:      m.inputs[0].Value(),
		Username: m.inputs[1].Value(),
		Password: m.inputs[2].Value(),
	}
}

// IsConnectionValid checks if the connection is valid
func (m ConnectionModel) IsConnectionValid() bool {
	return m.testSuccessful
}
