package ui

import (
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tehbooom/elastic-data/internal/integrations"
)

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

const (
	stateSelectingIntegration = iota
	stateSelectingDatasets
	stateConfiguringDataset
)

type IntegrationItem struct {
	title    string
	selected bool
}

type DatasetItem struct {
	name      string
	selected  bool
	threshold int
	unit      string
}

func (i DatasetItem) FilterValue() string {
	return i.name
}

func (i DatasetItem) Title() string {
	prefix := "  "
	if i.selected {
		prefix = "✓ "
	}
	return prefix + i.name
}

func (i DatasetItem) Description() string {
	return ""
}

func (i IntegrationItem) FilterValue() string {
	return i.title
}

func (i IntegrationItem) Title() string {
	prefix := "  "
	if i.selected {
		prefix = "✓ "
	}
	return prefix + i.title
}

func (i IntegrationItem) Description() string {
	return ""
}

// ToggleSelected toggles the selected state
func (i *IntegrationItem) ToggleSelected() {
	i.selected = !i.selected
}

type IntegrationsTabModel struct {
	width              int
	height             int
	integrationList    list.Model
	appState           *AppState
	state              int
	datasetsList       list.Model
	thresholdInput     textinput.Model
	unitInput          textinput.Model
	currentIntegration string
	viewport           viewport.Model
}

// CompactDelegate is a custom delegate with reduced spacing
type CompactDelegate struct {
	list.DefaultDelegate
}

// NewCompactDelegate creates a new compact delegate
func NewCompactDelegate() CompactDelegate {
	d := CompactDelegate{list.NewDefaultDelegate()}

	d.Styles.NormalTitle.
		PaddingLeft(0).
		MarginTop(0).
		PaddingTop(0).
		PaddingBottom(0)

	d.Styles.SelectedTitle.
		PaddingLeft(0).
		MarginTop(0).
		PaddingTop(0).
		PaddingBottom(0)

	d.Styles.NormalDesc.
		PaddingLeft(0).
		MarginTop(0).
		PaddingTop(0).
		PaddingBottom(0)

	d.Styles.SelectedDesc.
		PaddingLeft(0).
		MarginTop(0).
		PaddingTop(0).
		PaddingBottom(0)

	d.SetSpacing(0)
	d.ShowDescription = false
	return d
}

func NewIntegrationsTabModel(state *AppState) *IntegrationsTabModel {
	thInput := textinput.New()
	thInput.Placeholder = "Enter threshold value"
	thInput.CharLimit = 10
	// TODO: Add text validatioin for only ints

	uInput := textinput.New()
	uInput.Placeholder = "Unit (eps or bytes)"
	uInput.CharLimit = 5
	// TODO: Add text validation for eps or bytes

	vp := viewport.New(0, 0)

	delegate := NewCompactDelegate()
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Available Integrations"
	l.SetShowHelp(false)
	// TODO: Allow filtering
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)
	l.SetShowPagination(true)

	datasetsList := list.New([]list.Item{}, delegate, 0, 0)
	datasetsList.SetShowHelp(false)
	datasetsList.SetShowStatusBar(false)
	datasetsList.SetShowPagination(false)

	return &IntegrationsTabModel{
		integrationList: l,
		appState:        state,
		viewport:        vp,
		datasetsList:    datasetsList,
		thresholdInput:  thInput,
		unitInput:       uInput,
	}
}

func (m *IntegrationsTabModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	fmt.Printf("Setting IntegrationsTabModel size to width=%d, height=%d\n", width, height)

	m.integrationList.SetSize(width, height)
	m.datasetsList.SetSize(width, height)
}

func (m IntegrationsTabModel) TabTitle() string {
	return "Integrations"
}

func (m *IntegrationsTabModel) SetIntegrations(integrations []string) {
	m.integrationList.SetItems([]list.Item{})

	var items []list.Item
	for _, integration := range integrations {
		item := &IntegrationItem{
			title:    integration,
			selected: false,
		}
		items = append(items, item)

		if m.appState != nil && m.appState.selectedIntegrations != nil {
			m.appState.selectedIntegrations[integration] = false
		}
	}

	m.integrationList.SetItems(items)
}

// GetSelectedIntegrations returns a map of selected integrations
func (m *IntegrationsTabModel) GetSelectedIntegrations() map[string]bool {
	result := make(map[string]bool)
	for _, item := range m.integrationList.Items() {
		if i, ok := item.(*IntegrationItem); ok {
			result[i.title] = i.selected
		}
	}
	return result
}

func (m IntegrationsTabModel) Init() tea.Cmd {
	return nil
}

func (m *IntegrationsTabModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	// Check for global navigation keys first (tab switching)
	// We need to check if the message is a key message and if it's a tab-switching key
	// before we handle it in the state-specific code
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "tab", "right", "shift+tab", "left":
			// Let these keys bubble up to the TabsModel for tab switching
			// by returning without handling them
			if m.state != stateConfiguringDataset {
				// Only bubble up if we're not in configuration mode
				// In configuration mode, tab is used to switch between input fields
				return m, nil
			}
			// For configuration mode, we need special handling for shift+tab
			if keyMsg.String() == "shift+tab" && m.state == stateConfiguringDataset {
				if m.thresholdInput.Focused() {
					m.thresholdInput.Blur()
					m.unitInput.Focus()
					return m, nil
				} else {
					m.unitInput.Blur()
					m.thresholdInput.Focus()
					return m, nil
				}
			}
		}
	}

	// Handle different states
	switch m.state {
	case stateSelectingIntegration:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case " ": // Toggle selection on space key
				item, ok := m.integrationList.SelectedItem().(*IntegrationItem)
				if !ok {
					return m, nil
				}
				item.selected = !item.selected
				m.appState.selectedIntegrations[item.title] = item.selected
				items := m.integrationList.Items()
				items[m.integrationList.Index()] = item
				m.integrationList.SetItems(items)
				return m, nil
			case "enter":
				item, ok := m.integrationList.SelectedItem().(*IntegrationItem)
				if !ok {
					return m, nil
				}
				m.currentIntegration = item.title
				m.loadDatasetsForIntegration(item.title)
				m.state = stateSelectingDatasets
				return m, nil
			case "esc", "q":
				return m, tea.Quit
			}
		}
		// Pass other messages to integration list
		m.integrationList, cmd = m.integrationList.Update(msg)
		return m, cmd

	case stateSelectingDatasets:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case " ": // Toggle selection
				item, ok := m.datasetsList.SelectedItem().(DatasetItem)
				if !ok {
					return m, nil
				}
				item.selected = !item.selected
				items := m.datasetsList.Items()
				items[m.datasetsList.Index()] = item
				m.datasetsList.SetItems(items)
				m.updateDatasetConfigs()
				return m, nil
			case "enter":
				item, ok := m.datasetsList.SelectedItem().(DatasetItem)
				if !ok {
					return m, nil
				}
				if !item.selected {
					item.selected = true
					items := m.datasetsList.Items()
					items[m.datasetsList.Index()] = item
					m.datasetsList.SetItems(items)
					m.updateDatasetConfigs()
				}
				m.thresholdInput.SetValue(fmt.Sprintf("%d", item.threshold))
				m.thresholdInput.Focus()
				m.unitInput.SetValue(item.unit)
				m.state = stateConfiguringDataset
				return m, nil
			case "left", "esc", "q":
				m.state = stateSelectingIntegration
				return m, nil
			}
		}
		// Pass other messages to datasets list
		m.datasetsList, cmd = m.datasetsList.Update(msg)
		return m, cmd

	case stateConfiguringDataset:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "tab":
				if m.thresholdInput.Focused() {
					m.thresholdInput.Blur()
					m.unitInput.Focus()
				} else {
					m.unitInput.Blur()
					m.thresholdInput.Focus()
				}
				return m, nil
			case "enter":
				threshold, _ := strconv.Atoi(m.thresholdInput.Value())

				idx := m.datasetsList.Index()
				items := m.datasetsList.Items()
				item := items[idx].(DatasetItem)

				item.threshold = threshold
				item.unit = m.unitInput.Value()

				items[idx] = item
				m.datasetsList.SetItems(items)

				m.updateDatasetConfigs()

				m.state = stateSelectingDatasets
				return m, nil
			case "esc", "":
				m.state = stateSelectingDatasets
				return m, nil
			}
		}

		var cmd tea.Cmd
		if m.thresholdInput.Focused() {
			m.thresholdInput, cmd = m.thresholdInput.Update(msg)
		} else if m.unitInput.Focused() {
			m.unitInput, cmd = m.unitInput.Update(msg)
		}
		return m, cmd
	}

	return m, nil
}

func (m IntegrationsTabModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content strings.Builder

	// Render breadcrumb navigation
	breadcrumbs := ""
	if m.state > stateSelectingIntegration {
		breadcrumbs = fmt.Sprintf("%s > %s",
			breadcrumbStyle.Render("Integrations"),
			activecrumbStyle.Render(m.currentIntegration))

		if m.state == stateConfiguringDataset {
			item := m.datasetsList.SelectedItem().(DatasetItem)
			breadcrumbs = fmt.Sprintf("%s > %s > %s",
				breadcrumbStyle.Render("Integrations"),
				breadcrumbStyle.Render(m.currentIntegration),
				activecrumbStyle.Render(item.name))
		}
	}

	content.WriteString(breadcrumbs + "\n\n")

	// Render the current view based on state
	switch m.state {
	case stateSelectingIntegration:
		content.WriteString(m.integrationList.View())
		content.WriteString("\n\n")
		content.WriteString(helpStyle.Render("(space) Toggle selection, (enter) Configure datasets, (tab) Switch tabs, (q) Quit"))

	case stateSelectingDatasets:
		content.WriteString(m.datasetsList.View())
		content.WriteString("\n\n")

		content.WriteString("\n")
		content.WriteString(helpStyle.Render("(space) Toggle selection, (enter) Configure selected, (left) Back to integrations, (tab) Switch tabs, (q) Back"))

	case stateConfiguringDataset:
		// Show the configuration form
		item := m.datasetsList.SelectedItem().(DatasetItem)

		form := strings.Builder{}
		form.WriteString(fmt.Sprintf("\n  Configuring: %s\n\n", item.name))
		form.WriteString(fmt.Sprintf("  Threshold: %s\n", m.thresholdInput.View()))
		form.WriteString(fmt.Sprintf("  Unit: %s\n\n", m.unitInput.View()))
		form.WriteString(helpStyle.Render("  (enter) Save, (esc) Cancel, (tab) Switch fields"))

		content.WriteString(form.String())
	}
	return content.String()
}

// loadDatasetsForIntegration loads the datasets for the selected integration
func (m *IntegrationsTabModel) loadDatasetsForIntegration(integration string) {
	var datasetItems []list.Item

	configs, exists := m.appState.datasetConfigs[integration]
	if !exists {
		configDir, _ := getConfigDir()
		repoDir := filepath.Join(configDir, "integrations")
		dataSets, err := integrations.GetDatasets(repoDir, integration)
		if err != nil {
			log.Fatal(err)
		}

		configs = []DatasetConfig{}
		for _, ds := range dataSets {
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
		datasetItems = append(datasetItems, DatasetItem{
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
func (m *IntegrationsTabModel) updateDatasetConfigs() {
	var configs []DatasetConfig

	for _, item := range m.datasetsList.Items() {
		datasetItem, ok := item.(DatasetItem)
		if !ok {
			continue
		}

		config := DatasetConfig{
			Name:      datasetItem.name,
			Selected:  datasetItem.selected,
			Threshold: datasetItem.threshold,
			Unit:      datasetItem.unit,
		}
		configs = append(configs, config)
	}

	for i, c := range configs {
		fmt.Printf("  Config %d: %s (selected: %v, threshold: %d, unit: %s)\n",
			i, c.Name, c.Selected, c.Threshold, c.Unit)
	}

	if m.appState != nil && m.currentIntegration != "" {
		m.appState.datasetConfigs[m.currentIntegration] = configs
	} else {
		fmt.Printf("ERROR: Cannot update app state - appState: %v, currentIntegration: %s\n",
			m.appState != nil, m.currentIntegration)
	}
}
