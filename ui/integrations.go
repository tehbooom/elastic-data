// ui/integrations.go
package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tehbooom/elastic-data/ui/integration"
	"github.com/tehbooom/elastic-data/ui/state"
)

// IntegrationsTabModel is the model for the integrations tab
type IntegrationsTabModel struct {
	tabModel *integration.TabModel
}

// NewIntegrationsTabModel creates a new integrations tab model
func NewIntegrationsTabModel(state *state.AppState, saveController *state.SaveController) *IntegrationsTabModel {
	return &IntegrationsTabModel{
		tabModel: integration.NewTabModel(state, saveController),
	}
}

// Implement TabModel interface
func (m *IntegrationsTabModel) TabTitle() string {
	return m.tabModel.TabTitle()
}

func (m *IntegrationsTabModel) SetSize(width, height int) {
	m.tabModel.SetSize(width, height)
}

func (m *IntegrationsTabModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := m.tabModel.Update(msg)

	// If the update returns a TabModel, update our inner model
	if updatedModel, ok := model.(*integration.TabModel); ok {
		m.tabModel = updatedModel
		return m, cmd
	}

	// Fallback
	return m, cmd
}

func (m IntegrationsTabModel) View() string {
	return m.tabModel.View()
}

func (m IntegrationsTabModel) Init() tea.Cmd {
	return m.tabModel.Init()
}

// IsInConfigurationState returns true if the tab is in configuration state
func (m *IntegrationsTabModel) IsInConfigurationState() bool {
	return m.tabModel.IsInConfigurationState()
}

func (m *IntegrationsTabModel) SetIntegrations(integrations []string) {
	m.tabModel.SetIntegrations(integrations)
}
