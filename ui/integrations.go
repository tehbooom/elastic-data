package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type IntegrationItem struct {
	title    string
	selected bool
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
	width  int
	height int
	list   list.Model
	state  *AppState
}

func NewIntegrationsTabModel(state *AppState) *IntegrationsTabModel {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Available Integrations"
	l.SetShowHelp(true)
	// TODO: Allow filtering
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)
	l.SetShowPagination(true)

	return &IntegrationsTabModel{
		list:  l,
		state: state,
	}
}

func (m *IntegrationsTabModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	fmt.Printf("Setting IntegrationsTabModel size to width=%d, height=%d\n", width, height)

	m.list.SetSize(width, height)
}

func (m IntegrationsTabModel) TabTitle() string {
	return "Integrations"
}

func (m *IntegrationsTabModel) SetIntegrations(integrations []string) {
	m.list.SetItems([]list.Item{})

	var items []list.Item
	for _, integration := range integrations {
		item := &IntegrationItem{
			title:    integration,
			selected: false,
		}
		items = append(items, item)

		if m.state != nil && m.state.selectedIntegrations != nil {
			m.state.selectedIntegrations[integration] = false
		}
	}

	m.list.SetItems(items)
}

// GetSelectedIntegrations returns a map of selected integrations
func (m *IntegrationsTabModel) GetSelectedIntegrations() map[string]bool {
	result := make(map[string]bool)
	for _, item := range m.list.Items() {
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case " ": // Toggle selection on space key
			idx := m.list.Index()
			items := m.list.Items()
			if idx >= 0 && idx < len(items) {
				if i, ok := items[idx].(*IntegrationItem); ok {
					i.selected = !i.selected
					m.state.selectedIntegrations[i.title] = i.selected

					// Initialize dataset config if newly selected
					if i.selected && m.state.datasetConfigs != nil {
						if _, exists := m.state.datasetConfigs[i.title]; !exists {
							m.state.datasetConfigs[i.title] = []DatasetConfig{}
						}
					}

					// Important: update the list items
					m.list.SetItems(items)
				}
			}
			return m, nil
		}
	}

	// For ALL other key events, pass to the list
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m IntegrationsTabModel) View() string {
	return m.list.View()
}
