package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmModel represents the confirmation screen
type ConfirmModel struct {
	width     int
	height    int
	confirmed bool
	selected  int // 0 for Yes, 1 for No
}

// NewConfirmModel creates a new confirmation model
func NewConfirmModel() ConfirmModel {
	return ConfirmModel{
		selected: 0, // Default to Yes
	}
}

// SetSize updates the size of the confirmation model
func (m *ConfirmModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Init initializes the confirmation model
func (m ConfirmModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m ConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			m.selected = 0 // Yes

		case "right", "l":
			m.selected = 1 // No

		case "enter":
			if m.selected == 0 {
				m.confirmed = true
			} else {
				// Go back to connection screen
				// In a real app, you might want to send a message
				// to go back to the previous screen
			}
		}
	}

	return m, nil
}

// View renders the confirmation screen
func (m ConfirmModel) View() string {
	if m.width == 0 {
		return "Loading confirmation screen..."
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width - 2).
		Align(lipgloss.Center)

	title := titleStyle.Render("Start Data Generation")

	// Message
	messageStyle := lipgloss.NewStyle().
		Padding(2, 0).
		Width(m.width - 2).
		Align(lipgloss.Center)

	message := messageStyle.Render("Connection established. Ready to start data generation?")

	// Buttons
	yesStyle := lipgloss.NewStyle().
		Padding(0, 3)

	noStyle := lipgloss.NewStyle().
		Padding(0, 3)

	if m.selected == 0 {
		yesStyle = yesStyle.
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7D56F4"))
	} else {
		noStyle = noStyle.
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7D56F4"))
	}

	yes := yesStyle.Render("Yes")
	no := noStyle.Render("No")

	buttons := lipgloss.JoinHorizontal(
		lipgloss.Center,
		yes,
		"    ",
		no,
	)

	buttonsWrap := lipgloss.NewStyle().
		Width(m.width - 2).
		Align(lipgloss.Center).
		Render(buttons)

	helpText := "• ←/→ to navigate • Enter to select •"
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Align(lipgloss.Center).
		Width(m.width - 2)

	help := helpStyle.Render(helpText)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"\n",
		message,
		"\n",
		buttonsWrap,
		"\n",
		help,
	)
}

// IsComplete returns whether confirmation is complete
func (m ConfirmModel) IsComplete() bool {
	return m.confirmed
}

// IsConfirmed returns whether the user confirmed
func (m ConfirmModel) IsConfirmed() bool {
	return m.confirmed
}
