package run

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tehbooom/elastic-data/ui/context"
)

type RunTabModel struct {
	TabModel *TabModel
}

func NewRunTabModel(context *context.ProgramContext, saveController *context.SaveController) *RunTabModel {
	return &RunTabModel{
		TabModel: NewTabModel(context, saveController),
	}
}

func (m *RunTabModel) TabTitle() string {
	return m.TabModel.TabTitle()
}

func (m *RunTabModel) SetSize(width, height int) {
	m.TabModel.SetSize(width, height)
}

func (m *RunTabModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := m.TabModel.Update(msg)

	if updatedModel, ok := model.(*TabModel); ok {
		m.TabModel = updatedModel
		return m, cmd
	}

	return m, cmd
}

func (m RunTabModel) View() string {
	return m.TabModel.View()
}

func (m RunTabModel) Init() tea.Cmd {
	return m.TabModel.Init()
}
