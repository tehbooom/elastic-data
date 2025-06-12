package integration

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tehbooom/elastic-data/ui/context"
)

type IntegrationsTabModel struct {
	tabModel *TabModel
}

func NewIntegrationsTabModel(state *context.ProgramContext, saveController *context.SaveController) *IntegrationsTabModel {
	return &IntegrationsTabModel{
		tabModel: NewTabModel(state, saveController),
	}
}

func (m *IntegrationsTabModel) TabTitle() string {
	return m.tabModel.TabTitle()
}

func (m *IntegrationsTabModel) SetSize(width, height int) {
	m.tabModel.SetSize(width, height)
}

func (m *IntegrationsTabModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := m.tabModel.Update(msg)

	if updatedModel, ok := model.(*TabModel); ok {
		m.tabModel = updatedModel
		return m, cmd
	}

	return m, cmd
}

func (m IntegrationsTabModel) View() string {
	return m.tabModel.View()
}

func (m IntegrationsTabModel) Init() tea.Cmd {
	return m.tabModel.Init()
}

func (m *IntegrationsTabModel) IsInConfigurationState() bool {
	return m.tabModel.IsInConfigurationState()
}

func (m *IntegrationsTabModel) SetIntegrations(integrations []string) {
	m.tabModel.SetIntegrations(integrations)
}
