// ui/integrations.go
package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tehbooom/elastic-data/ui/run"
	"github.com/tehbooom/elastic-data/ui/state"
)

// RunTabModel is the model for the integrations tab
type RunTabModel struct {
	tabModel *run.TabModel
}

// NewRunTabModel creates a new integrations tab model
func NewRunTabModel(state *state.AppState, saveController *state.SaveController) *RunTabModel {
	return &RunTabModel{
		tabModel: run.NewTabModel(state, saveController),
	}
}

// Implement TabModel interface
func (m *RunTabModel) TabTitle() string {
	return m.tabModel.TabTitle()
}

func (m *RunTabModel) SetSize(width, height int) {
	m.tabModel.SetSize(width, height)
}

func (m *RunTabModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := m.tabModel.Update(msg)

	// If the update returns a TabModel, update our inner model
	if updatedModel, ok := model.(*run.TabModel); ok {
		m.tabModel = updatedModel
		return m, cmd
	}

	// Fallback
	return m, cmd
}

func (m RunTabModel) View() string {
	return m.tabModel.View()
}

func (m RunTabModel) Init() tea.Cmd {
	return m.tabModel.Init()
}
