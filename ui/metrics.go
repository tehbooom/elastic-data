package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MetricsModel represents the metrics selection screen
type MetricsModel struct {
	selectedDatasets map[string][]string // From previous screen
	width            int
	height           int
	complete         bool
	selectedMetrics  map[string]string // Map of datasetKey -> metric type ("total" or "eps")

	// For navigation
	currentItemIndex      int
	currentDatasetIndices map[string]int
	flattenedDatasets     []datasetKey // For easier navigation
}

// datasetKey uniquely identifies a dataset
type datasetKey struct {
	item    string
	dataset string
}

// String returns a string representation of the dataset key
func (dk datasetKey) String() string {
	return fmt.Sprintf("%s: %s", dk.item, dk.dataset)
}

// NewMetricsModel creates a new metrics model
func NewMetricsModel() MetricsModel {
	return MetricsModel{
		selectedMetrics:       make(map[string]string),
		currentDatasetIndices: make(map[string]int),
	}
}

// SetSize updates the size of the metrics model
func (m *MetricsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetSelectedDatasets sets the datasets for which metrics will be selected
func (m *MetricsModel) SetSelectedDatasets(datasets map[string][]string) {
	m.selectedDatasets = datasets
	m.flattenedDatasets = flattenDatasets(datasets)
	m.currentItemIndex = 0

	// Initialize default metric selection
	for _, dk := range m.flattenedDatasets {
		keyStr := dk.String()
		m.selectedMetrics[keyStr] = "total" // Default to total
	}
}

// flattenDatasets converts the nested datasets map to a flat slice for easier navigation
func flattenDatasets(datasets map[string][]string) []datasetKey {
	var flattened []datasetKey

	for item, itemDatasets := range datasets {
		for _, dataset := range itemDatasets {
			flattened = append(flattened, datasetKey{item: item, dataset: dataset})
		}
	}

	return flattened
}

// Init initializes the metrics model
func (m MetricsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m MetricsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if len(m.flattenedDatasets) == 0 {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down", "j", "tab": // Next dataset
			if m.currentItemIndex < len(m.flattenedDatasets)-1 {
				m.currentItemIndex++
			}

		case "up", "k", "shift+tab": // Previous dataset
			if m.currentItemIndex > 0 {
				m.currentItemIndex--
			}

		case "left", "h": // Select total
			currKey := m.flattenedDatasets[m.currentItemIndex].String()
			m.selectedMetrics[currKey] = "total"

		case "right", "l": // Select EPS
			currKey := m.flattenedDatasets[m.currentItemIndex].String()
			m.selectedMetrics[currKey] = "eps"

		case "enter": // Confirm selections
			m.complete = true
		}
	}

	return m, nil
}

// View renders the metrics selection screen
func (m MetricsModel) View() string {
	if len(m.flattenedDatasets) == 0 {
		return "No datasets selected..."
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width - 2).
		Align(lipgloss.Center)

	title := titleStyle.Render("Select Metrics Type")

	// Build the table of options
	rows := []string{}

	for i, dk := range m.flattenedDatasets {
		keyStr := dk.String()
		metric := m.selectedMetrics[keyStr]

		// Highlight the current row
		rowStyle := lipgloss.NewStyle().Width(m.width - 4)
		if i == m.currentItemIndex {
			rowStyle = rowStyle.Background(lipgloss.Color("#2D2D2D"))
		}

		// Build option styles
		totalStyle := lipgloss.NewStyle().Padding(0, 1)
		epsStyle := lipgloss.NewStyle().Padding(0, 1)

		if metric == "total" {
			totalStyle = totalStyle.
				Background(lipgloss.Color("#7D56F4")).
				Foreground(lipgloss.Color("#FFFFFF"))
		} else {
			epsStyle = epsStyle.
				Background(lipgloss.Color("#7D56F4")).
				Foreground(lipgloss.Color("#FFFFFF"))
		}

		total := totalStyle.Render("Total Data Sent")
		eps := epsStyle.Render("Events Per Second")

		datasetLabel := lipgloss.NewStyle().
			Width(25).
			Align(lipgloss.Left).
			Render(keyStr)

		options := lipgloss.JoinHorizontal(
			lipgloss.Center,
			datasetLabel,
			"  ",
			total,
			"  ",
			eps,
		)

		row := rowStyle.Render(options)
		rows = append(rows, row)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)

	helpText := "• ↑/↓ to navigate • ←/→ to select metric • Enter to confirm •"
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Align(lipgloss.Center).
		Width(m.width - 2)

	help := helpStyle.Render(helpText)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		content,
		help,
	)
}

// IsComplete returns whether metrics selection is complete
func (m MetricsModel) IsComplete() bool {
	return m.complete
}

// GetSelectedMetrics returns the map of selected metrics
func (m MetricsModel) GetSelectedMetrics() map[string]string {
	return m.selectedMetrics
}
