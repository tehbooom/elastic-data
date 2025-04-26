package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SelectionModel represents the item selection screen
type SelectionModel struct {
	list     list.Model
	width    int
	height   int
	title    string
	complete bool
}

// item represents a selectable item
type item struct {
	title       string
	description string
	selected    bool
}

// FilterValue implements list.Item interface
func (i item) FilterValue() string {
	return i.title
}

// Title returns the title of the item
func (i item) Title() string {
	prefix := "  "
	if i.selected {
		prefix = "✓ "
	}
	return prefix + i.title
}

// Description returns the description of the item
func (i item) Description() string {
	return i.description
}

// NewSelectionModel creates a new selection model
func NewSelectionModel() SelectionModel {
	// Sample items - in a real app, these would come from your data source
	items := []list.Item{
		item{title: "Elastic Common Schema", description: "ECS compliant logs", selected: false},
		item{title: "Winlogbeat", description: "Windows event logs", selected: false},
		item{title: "Nginx", description: "Web server logs", selected: false},
		item{title: "Apache", description: "Web server logs", selected: false},
		item{title: "Kubernetes", description: "Container orchestration logs", selected: false},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select Data Types"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return SelectionModel{
		list:  l,
		title: "Select one or more data types to generate",
	}
}

// SetSize updates the size of the selection model
func (m *SelectionModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	m.list.SetSize(width, height-4) // Subtract space for title and help
}

// Init initializes the selection model
func (m SelectionModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m SelectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case " ": // Toggle selection
			idx := m.list.Index()
			items := m.list.Items()
			if idx >= 0 && idx < len(items) {
				i := items[idx].(item)
				i.selected = !i.selected
				items[idx] = i
				m.list.SetItems(items)
			}
			return m, nil

		case "enter": // Confirm selection
			// Check if at least one item is selected
			for _, i := range m.list.Items() {
				if i.(item).selected {
					m.complete = true
					break
				}
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the selection screen
func (m SelectionModel) View() string {
	if m.width == 0 {
		return "Loading selection screen..."
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width - 2).
		Align(lipgloss.Center)

	title := titleStyle.Render(m.title)

	helpText := "• Space to select/deselect • Enter to confirm •"
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Align(lipgloss.Center).
		Width(m.width - 2)

	help := helpStyle.Render(helpText)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		m.list.View(),
		help,
	)
}

// IsComplete returns whether selection is complete
func (m SelectionModel) IsComplete() bool {
	return m.complete
}

// GetSelectedItems returns the list of selected items
func (m SelectionModel) GetSelectedItems() []string {
	var selectedItems []string

	for _, i := range m.list.Items() {
		item := i.(item)
		if item.selected {
			selectedItems = append(selectedItems, item.title)
		}
	}

	return selectedItems
}
