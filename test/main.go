package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AppState holds the shared state between tabs
type AppState struct {
	selectedIntegrations map[string]bool
	datasetConfigs       map[string][]DatasetConfig
}

// Create a new AppState
func NewAppState() *AppState {
	return &AppState{
		selectedIntegrations: make(map[string]bool),
		datasetConfigs:       make(map[string][]DatasetConfig),
	}
}

type DatasetConfig struct {
	Name      string
	Selected  bool
	Threshold int
	// EPS or bytes
	Unit string
}

// UI states
const (
	stateSelectingIntegration = iota
	stateSelectingDatasets
	stateConfiguringDataset
)

// Define styles
var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Bold(true)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	checkboxStyle     = lipgloss.NewStyle().PaddingRight(1)
	infoStyle         = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("240"))
	configStyle       = lipgloss.NewStyle().PaddingLeft(6).Foreground(lipgloss.Color("132"))
	breadcrumbStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("110"))
	activecrumbStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
	helpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

// IntegrationModel manages both integration selection and dataset configuration
type IntegrationModel struct {
	appState           *AppState
	state              int
	integrationsList   list.Model
	datasetsList       list.Model
	currentIntegration string
	thresholdInput     textinput.Model
	unitInput          textinput.Model
	viewport           viewport.Model
	width              int
	height             int
}

// integrationItem represents an integration in the list
type integrationItem struct {
	name     string
	selected bool
}

func (i integrationItem) FilterValue() string { return i.name }

// datasetItem represents a dataset in the list
type datasetItem struct {
	name      string
	selected  bool
	threshold int
	unit      string
}

func (i datasetItem) FilterValue() string { return i.name }

// integrationDelegate defines how integration items are rendered
type integrationDelegate struct{}

func (d integrationDelegate) Height() int                               { return 1 }
func (d integrationDelegate) Spacing() int                              { return 0 }
func (d integrationDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d integrationDelegate) Render(w int, m list.Model, index int, listItem list.Item) string {
	item, ok := listItem.(integrationItem)
	if !ok {
		return "Error: item is not an integrationItem"
	}

	checkbox := "[ ]"
	if item.selected {
		checkbox = "[x]"
	}

	str := fmt.Sprintf("%s %s", checkboxStyle.Render(checkbox), item.name)

	if index == m.Index() {
		return selectedItemStyle.Render("> " + str)
	}

	return itemStyle.Render(str)
}

// datasetDelegate defines how dataset items are rendered
type datasetDelegate struct{}

func (d datasetDelegate) Height() int                               { return 1 }
func (d datasetDelegate) Spacing() int                              { return 0 }
func (d datasetDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d datasetDelegate) Render(w int, m list.Model, index int, listItem list.Item) string {
	item, ok := listItem.(datasetItem)
	if !ok {
		return "Error: item is not a datasetItem"
	}

	checkbox := "[ ]"
	if item.selected {
		checkbox = "[x]"
	}

	str := fmt.Sprintf("%s %s", checkboxStyle.Render(checkbox), item.name)

	if index == m.Index() {
		return selectedItemStyle.Render("> " + str)
	}

	return itemStyle.Render(str)
}

// NewIntegrationModel creates a new integration configuration model
func NewIntegrationModel(appState *AppState) *IntegrationModel {
	// Create threshold input
	thInput := textinput.New()
	thInput.Placeholder = "Enter threshold value"
	thInput.CharLimit = 10

	// Create unit input
	uInput := textinput.New()
	uInput.Placeholder = "Unit (eps or bytes)"
	uInput.CharLimit = 5

	// Create viewport for scrolling
	vp := viewport.New(0, 0)
	vp.HighPerformanceRendering = true

	// Convert integrations to list items
	var integrationItems []list.Item

	// Sample integrations - in real app, get these from somewhere
	integrations := []string{"aws", "gcp", "azure", "datadog", "prometheus", "elasticsearch",
		"kafka", "rabbitmq", "redis", "mysql", "postgresql", "mongodb"}

	for _, integration := range integrations {
		selected := false
		if _, exists := appState.selectedIntegrations[integration]; exists {
			selected = appState.selectedIntegrations[integration]
		}

		integrationItems = append(integrationItems, integrationItem{
			name:     integration,
			selected: selected,
		})
	}

	// Create integrations list
	integrationDelegate := integrationDelegate{}
	intList := list.New(integrationItems, integrationDelegate, 0, 0)
	intList.Title = "Available Integrations"
	intList.SetShowStatusBar(false)
	intList.SetFilteringEnabled(false)
	intList.SetShowHelp(false)

	// Create empty datasets list (will be populated when integration is selected)
	datasetDelegate := datasetDelegate{}
	datasetList := list.New([]list.Item{}, datasetDelegate, 0, 0)
	datasetList.SetShowStatusBar(false)
	datasetList.SetFilteringEnabled(false)
	datasetList.SetShowHelp(false)

	return &IntegrationModel{
		appState:         appState,
		state:            stateSelectingIntegration,
		integrationsList: intList,
		datasetsList:     datasetList,
		thresholdInput:   thInput,
		unitInput:        uInput,
		viewport:         vp,
	}
}

func (m *IntegrationModel) Init() tea.Cmd {
	return nil
}

func (m *IntegrationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 4
		footerHeight := 3
		availableHeight := m.height - headerHeight - footerHeight

		m.integrationsList.SetSize(msg.Width, availableHeight)
		m.datasetsList.SetSize(msg.Width, availableHeight)
		m.viewport.Width = msg.Width
		m.viewport.Height = availableHeight

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			if m.state == stateConfiguringDataset {
				// Exit configuration mode
				m.state = stateSelectingDatasets
				return m, nil
			} else if m.state == stateSelectingDatasets {
				// Go back to integration selection
				m.state = stateSelectingIntegration
				return m, nil
			} else {
				return m, tea.Quit
			}

		case "enter":
			switch m.state {
			case stateSelectingIntegration:
				// Get selected integration
				item := m.integrationsList.SelectedItem().(integrationItem)

				// Toggle selection
				item.selected = !item.selected
				m.appState.selectedIntegrations[item.name] = item.selected

				// Update the list
				items := m.integrationsList.Items()
				items[m.integrationsList.Index()] = item
				m.integrationsList.SetItems(items)

				// If selected, go to dataset selection
				if item.selected {
					m.currentIntegration = item.name
					m.loadDatasetsForIntegration(item.name)
					m.state = stateSelectingDatasets
				}

			case stateSelectingDatasets:
				// Get selected dataset
				item := m.datasetsList.SelectedItem().(datasetItem)

				// Toggle selection
				item.selected = !item.selected

				// If selected, go to configuration mode
				if item.selected {
					// Set input values for configuration
					m.thresholdInput.SetValue(fmt.Sprintf("%d", item.threshold))
					m.thresholdInput.Focus()
					m.unitInput.SetValue(item.unit)
					m.state = stateConfiguringDataset
				}

				// Update the list
				items := m.datasetsList.Items()
				items[m.datasetsList.Index()] = item
				m.datasetsList.SetItems(items)
				m.updateDatasetConfigs()

			case stateConfiguringDataset:
				// Save the configuration
				item := m.datasetsList.SelectedItem().(datasetItem)

				// Parse the threshold value
				var threshold int
				fmt.Sscanf(m.thresholdInput.Value(), "%d", &threshold)

				// Update the item
				item.threshold = threshold
				item.unit = m.unitInput.Value()

				// Update the list
				items := m.datasetsList.Items()
				items[m.datasetsList.Index()] = item
				m.datasetsList.SetItems(items)

				// Update the configs
				m.updateDatasetConfigs()

				// Exit configuration mode
				m.state = stateSelectingDatasets
				m.thresholdInput.Blur()
				m.unitInput.Blur()
			}

		case "space":
			switch m.state {
			case stateSelectingIntegration:
				// Toggle integration selection
				item := m.integrationsList.SelectedItem().(integrationItem)
				item.selected = !item.selected
				m.appState.selectedIntegrations[item.name] = item.selected

				// Update the list
				items := m.integrationsList.Items()
				items[m.integrationsList.Index()] = item
				m.integrationsList.SetItems(items)

			case stateSelectingDatasets:
				// Toggle dataset selection
				item := m.datasetsList.SelectedItem().(datasetItem)
				item.selected = !item.selected

				// Update the list
				items := m.datasetsList.Items()
				items[m.datasetsList.Index()] = item
				m.datasetsList.SetItems(items)
				m.updateDatasetConfigs()
			}

		case "tab":
			if m.state == stateConfiguringDataset {
				if m.thresholdInput.Focused() {
					m.thresholdInput.Blur()
					m.unitInput.Focus()
				} else {
					m.unitInput.Blur()
					m.thresholdInput.Focus()
				}
			}

		case "right", "l":
			if m.state == stateSelectingIntegration {
				// If the selected integration is selected, go to dataset selection
				item := m.integrationsList.SelectedItem().(integrationItem)
				if item.selected {
					m.currentIntegration = item.name
					m.loadDatasetsForIntegration(item.name)
					m.state = stateSelectingDatasets
				}
			}

		case "left", "h":
			if m.state == stateSelectingDatasets {
				// Go back to integration selection
				m.state = stateSelectingIntegration
			}
		}
	}

	// Handle input updates when in configuration mode
	if m.state == stateConfiguringDataset {
		if m.thresholdInput.Focused() {
			m.thresholdInput, cmd = m.thresholdInput.Update(msg)
			return m, cmd
		} else if m.unitInput.Focused() {
			m.unitInput, cmd = m.unitInput.Update(msg)
			return m, cmd
		}
	}

	// Handle list updates based on current state
	switch m.state {
	case stateSelectingIntegration:
		m.integrationsList, cmd = m.integrationsList.Update(msg)
		cmds = append(cmds, cmd)

	case stateSelectingDatasets:
		m.datasetsList, cmd = m.datasetsList.Update(msg)
		cmds = append(cmds, cmd)

	case stateConfiguringDataset:
		// Already handled above
	}

	// Update viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// loadDatasetsForIntegration loads the datasets for the selected integration
func (m *IntegrationModel) loadDatasetsForIntegration(integration string) {
	var datasetItems []list.Item

	// Get existing configs or create default ones
	configs, exists := m.appState.datasetConfigs[integration]
	if !exists {
		// Sample datasets - in real app, get these from somewhere
		sampleDatasets := []string{"cpu_usage", "memory_usage", "disk_space", "network_traffic",
			"connections", "requests", "errors", "latency"}

		configs = []DatasetConfig{}
		for _, ds := range sampleDatasets {
			configs = append(configs, DatasetConfig{
				Name:      ds,
				Selected:  false,
				Threshold: 0,
				Unit:      "eps",
			})
		}

		m.appState.datasetConfigs[integration] = configs
	}

	// Convert configs to list items
	for _, config := range configs {
		datasetItems = append(datasetItems, datasetItem{
			name:      config.Name,
			selected:  config.Selected,
			threshold: config.Threshold,
			unit:      config.Unit,
		})
	}

	// Update the datasets list
	m.datasetsList.SetItems(datasetItems)
	m.datasetsList.Title = fmt.Sprintf("%s Datasets", strings.ToUpper(integration))
}

// updateDatasetConfigs updates the app state with the current dataset configurations
func (m *IntegrationModel) updateDatasetConfigs() {
	var configs []DatasetConfig

	for _, item := range m.datasetsList.Items() {
		datasetItem := item.(datasetItem)
		config := DatasetConfig{
			Name:      datasetItem.name,
			Selected:  datasetItem.selected,
			Threshold: datasetItem.threshold,
			Unit:      datasetItem.unit,
		}
		configs = append(configs, config)
	}

	// Update the app state
	m.appState.datasetConfigs[m.currentIntegration] = configs
}

// View renders the current state
func (m *IntegrationModel) View() string {
	var content strings.Builder

	// Render breadcrumb navigation
	breadcrumbs := "Integrations"
	if m.state > stateSelectingIntegration {
		breadcrumbs = fmt.Sprintf("%s > %s",
			breadcrumbStyle.Render("Integrations"),
			activecrumbStyle.Render(m.currentIntegration))

		if m.state == stateConfiguringDataset {
			item := m.datasetsList.SelectedItem().(datasetItem)
			breadcrumbs = fmt.Sprintf("%s > %s > %s",
				breadcrumbStyle.Render("Integrations"),
				breadcrumbStyle.Render(m.currentIntegration),
				activecrumbStyle.Render(item.name))
		}
	} else {
		breadcrumbs = activecrumbStyle.Render("Integrations")
	}

	content.WriteString(breadcrumbs + "\n\n")

	// Render the current view based on state
	switch m.state {
	case stateSelectingIntegration:
		content.WriteString(m.integrationsList.View())
		content.WriteString("\n\n")
		content.WriteString(helpStyle.Render("(space) Toggle selection, (right/enter) Configure datasets, (q) Quit"))

	case stateSelectingDatasets:
		content.WriteString(m.datasetsList.View())
		content.WriteString("\n\n")

		// Show selected datasets with configuration
		content.WriteString("  Selected Datasets:\n")
		hasSelected := false

		for _, item := range m.datasetsList.Items() {
			datasetItem := item.(datasetItem)
			if datasetItem.selected {
				hasSelected = true
				configLine := fmt.Sprintf("%s: Threshold: %d, Unit: %s",
					datasetItem.name, datasetItem.threshold, datasetItem.unit)
				content.WriteString(configStyle.Render(configLine) + "\n")
			}
		}

		if !hasSelected {
			content.WriteString(infoStyle.Render("  No datasets selected.\n"))
		}

		content.WriteString("\n")
		content.WriteString(helpStyle.Render("(space) Toggle selection, (enter) Configure selected, (left) Back to integrations, (q) Back"))

	case stateConfiguringDataset:
		// Show the configuration form
		item := m.datasetsList.SelectedItem().(datasetItem)

		form := strings.Builder{}
		form.WriteString(fmt.Sprintf("\n  Configuring: %s\n\n", item.name))
		form.WriteString(fmt.Sprintf("  Threshold: %s\n", m.thresholdInput.View()))
		form.WriteString(fmt.Sprintf("  Unit: %s\n\n", m.unitInput.View()))
		form.WriteString(helpStyle.Render("  (Enter) Save, (Esc) Cancel, (Tab) Switch fields"))

		content.WriteString(form.String())
	}

	return content.String()
}

// RunIntegrationSelector initializes and runs the integration selector
func RunIntegrationSelector() {
	appState := NewAppState()
	model := NewIntegrationModel(appState)
	p := tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
