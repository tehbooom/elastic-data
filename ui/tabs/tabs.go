package tabs

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	ProgramContext "github.com/tehbooom/elastic-data/ui/context"
	"github.com/tehbooom/elastic-data/ui/errors"
	"github.com/tehbooom/elastic-data/ui/tabs/integration"
	"github.com/tehbooom/elastic-data/ui/tabs/run"
)

type TabModel interface {
	tea.Model
	SetSize(width, height int)
	TabTitle() string
}

type TabsModel struct {
	Tabs           []TabModel
	programContext *ProgramContext.ProgramContext
	ActiveTab      int
	Width          int
	Height         int
}

func NewTabsModel(tabs []TabModel, ctx *ProgramContext.ProgramContext) TabsModel {
	return TabsModel{
		Tabs:           tabs,
		programContext: ctx,
		ActiveTab:      0,
	}
}

func (m *TabsModel) SetSize(width, height int) {
	m.Width = width
	m.Height = height
	tabHeight := height - 6
	for i := range m.Tabs {
		m.Tabs[i].SetSize(width, tabHeight)
	}
}

func (m TabsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case errors.ShowErrorMsg:
		return m, func() tea.Msg { return msg }
	case tea.KeyMsg:
		if integrationsTab, ok := m.Tabs[m.ActiveTab].(*integration.IntegrationsTabModel); ok {
			if integrationsTab.IsInConfigurationState() {
				tabModel, cmd := integrationsTab.Update(msg)

				if updatedTab, ok := tabModel.(TabModel); ok {
					m.Tabs[m.ActiveTab] = updatedTab
				}
				return m, cmd
			}
		}

		switch msg.String() {
		case "tab":
			m.ActiveTab = (m.ActiveTab + 1) % len(m.Tabs)

			if runTab, ok := m.Tabs[m.ActiveTab].(*run.RunTabModel); ok {
				runTab.TabModel.RefreshIntegrations()
				if m.programContext.IsRunning() {
					return m, run.CreateTickCmd()
				}
			}

			return m, nil
		}
	}

	if m.ActiveTab < len(m.Tabs) {
		tabModel, cmd := m.Tabs[m.ActiveTab].Update(msg)
		if updatedTab, ok := tabModel.(TabModel); ok {
			m.Tabs[m.ActiveTab] = updatedTab
		}
		return m, cmd
	}

	return m, nil
}

func (m TabsModel) View() string {
	var header strings.Builder
	for i, tab := range m.Tabs {
		if i == m.ActiveTab {
			header.WriteString(lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#0B64DD")).
				Padding(0, 1).
				Render("[ " + tab.TabTitle() + " ]"))
		} else {
			header.WriteString(lipgloss.NewStyle().
				Padding(0, 1).
				Render("  " + tab.TabTitle() + "  "))
		}
	}

	content := m.Tabs[m.ActiveTab].View()

	return lipgloss.JoinVertical(lipgloss.Left, header.String(), content)
}

func (m TabsModel) Init() tea.Cmd {
	return nil
}
