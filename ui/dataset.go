package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DatasetModel represents the dataset selection screen
type DatasetModel struct {
	items        []string     // Selected items from previous screen
	currentItem  int          // Index of the current item
	lists        []list.Model // One list per item
	width        int
	height       int
	complete     bool
	selectedData map[string][]string // Map of item -> selected datasets
}

// datasetItem represents a selectable dataset
type datasetItem struct {
	title       string
	description string
	selected    bool
}

// FilterValue implements list.Item interface
func (i datasetItem) FilterValue() string {
	return i.title
}

// Title returns the title of the dataset
func (i datasetItem) Title() string {
	prefix := "  "
	if i.selected {
		prefix = "✓ "
	}
	return prefix + i.title
}

// Description returns the description of the dataset
func (i datasetItem) Description() string {
	return i.description
}

// NewDatasetModel creates a new dataset model
func NewDatasetModel() DatasetModel {
	return DatasetModel{
		selectedData: make(map[string][]string),
	}
}

// SetSize updates the size of the dataset model
func (m *DatasetModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Update size for all lists
	for i := range m.lists {
		m.lists[i].SetSize(width, height-8) // Space for title, help, and navigation
	}
}

// SetAvailableItems sets the items for which datasets will be selected
func (m *DatasetModel) SetAvailableItems(items []string) {
	m.items = items
	m.currentItem = 0
	m.lists = make([]list.Model, len(items))

	// Create a list for each item
	for i, item := range items {
		// Get datasets for this item - in real app, this would come from your data source
		datasets := getDatasetsByItem(item)

		l := list.New(datasets, list.NewDefaultDelegate(), 0, 0)
		l.Title = fmt.Sprintf("Select datasets for %s", item)
		l.SetShowHelp(false)
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(false)

		m.lists[i] = l
	}
}

// getDatasetsByItem returns datasets available for a given item
// In a real app, this would query your data source
func getDatasetsByItem(item string) []list.Item {
	var datasets []list.Item

	switch item {
	case "Elastic Common Schema":
		datasets = []list.Item{
			datasetItem{title: "Authentication", description: "Auth logs", selected: false},
			datasetItem{title: "Network", description: "Network logs", selected: false},
			datasetItem{title: "Process", description: "Process logs", selected: false},
		}
	case "Winlogbeat":
		datasets = []list.Item{
			datasetItem{title: "Security", description: "Security logs", selected: false},
			datasetItem{title: "System", description: "System logs", selected: false},
			datasetItem{title: "Application", description: "Application logs", selected: false},
		}
	case "Nginx":
		datasets = []list.Item{
			datasetItem{title: "Access", description: "Access logs", selected: false},
			datasetItem{title: "Error", description: "Error logs", selected: false},
		}
	case "Apache":
		datasets = []list.Item{
			datasetItem{title: "Access", description: "Access logs", selected: false},
			datasetItem{title: "Error", description: "Error logs", selected: false},
		}
	case "Kubernetes":
		datasets = []list.Item{
			datasetItem{title: "Pod", description: "Pod logs", selected: false},
			datasetItem{title: "Container", description: "Container logs", selected: false},
			datasetItem{title: "Node", description: "Node logs", selected: false},
		}
	default:
		datasets = []list.Item{
			datasetItem{title: "Default", description: "Default dataset", selected: false},
		}
	}

	return datasets
}

// Init initializes the dataset model
func (m DatasetModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m DatasetModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if len(m.items) == 0 {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case " ": // Toggle selection
			if m.currentItem >= 0 && m.currentItem < len(m.lists) {
				list := m.lists[m.currentItem]
				idx := list.Index()
				items := list.Items()

				if idx >= 0 && idx < len(items) {
					i := items[idx].(datasetItem)
					i.selected = !i.selected
					items[idx] = i
					list.SetItems(items)
					m.lists[m.currentItem] = list
				}
			}

		case "tab", "right", "l": // Next item
			if m.currentItem < len(m.items)-1 {
				m.currentItem++
			}

		case "shift+tab", "left", "h": // Previous item
			if m.currentItem > 0 {
				m.currentItem--
			}

		case "enter": // Confirm selection
			// Save selected datasets
			for i, item := range m.items {
				var selected []string
				for _, di := range m.lists[i].Items() {
					d := di.(datasetItem)
					if d.selected {
						selected = append(selected, d.title)
					}
				}

				// Only save if at least one dataset is selected
				if len(selected) > 0 {
					m.selectedData[item] = selected
				}
			}

			// Ensure at least one item has datasets selected
			if len(m.selectedData) > 0 {
				m.complete = true
			}
		}
	}

	// Update the current list
	if m.currentItem >= 0 && m.currentItem < len(m.lists) {
		var cmd tea.Cmd
		m.lists[m.currentItem], cmd = m.lists[m.currentItem].Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the dataset selection screen
func (m DatasetModel) View() string {
	if len(m.items) == 0 || m.currentItem >= len(m.lists) {
		return "No items selected or loading datasets..."
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width - 2).
		Align(lipgloss.Center)

	title := titleStyle.Render(fmt.Sprintf("Dataset Selection (%d/%d)", m.currentItem+1, len(m.items)))

	// Navigation help
	navHelp := fmt.Sprintf("• Item: %s •", m.items[m.currentItem])
	if len(m.items) > 1 {
		navButtons := ""
		if m.currentItem > 0 {
			navButtons += "◀ Prev "
		}
		if m.currentItem < len(m.items)-1 {
			navButtons += "Next ▶"
		}
		navHelp = fmt.Sprintf("• Item: %s • %s •", m.items[m.currentItem], navButtons)
	}

	navStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#383838")).
		Padding(0, 1).
		Width(m.width - 2).
		Align(lipgloss.Center)

	nav := navStyle.Render(navHelp)

	helpText := "• Space to select/deselect • Tab/Arrow to navigate • Enter to confirm •"
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Align(lipgloss.Center).
		Width(m.width - 2)

	help := helpStyle.Render(helpText)

	// Current list view
	listView := m.lists[m.currentItem].View()

	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		nav,
		listView,
		help,
	)
}

// IsComplete returns whether dataset selection is complete
func (m DatasetModel) IsComplete() bool {
	return m.complete
}

// GetSelectedDatasets returns the map of selected datasets
func (m DatasetModel) GetSelectedDatasets() map[string][]string {
	return m.selectedData
}
