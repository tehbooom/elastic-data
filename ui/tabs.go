package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TabsModel struct {
	tabs      []TabModel
	activeTab int
	width     int
	height    int
}

func NewTabsModel(tabs []TabModel) TabsModel {
	return TabsModel{
		tabs:      tabs,
		activeTab: 0,
	}
}

func (m *TabsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	tabHeight := height - 6
	for i := range m.tabs {
		m.tabs[i].SetSize(width, tabHeight)
	}
}

func (m TabsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Check if the current tab is in a configuration state
		if integrationsTab, ok := m.tabs[m.activeTab].(*IntegrationsTabModel); ok {
			if integrationsTab.IsInConfigurationState() {
				// When in configuration mode, let the tab handle the keys first
				tabModel, cmd := integrationsTab.Update(msg)

				// Update the tab in the slice if it implements TabModel
				if updatedTab, ok := tabModel.(TabModel); ok {
					m.tabs[m.activeTab] = updatedTab
				}
				return m, cmd
			}
		}

		// Handle tab switching keys if not in configuration state
		switch msg.String() {
		case "tab", "right":
			// Move to next tab
			m.activeTab = (m.activeTab + 1) % len(m.tabs)
			return m, nil
		case "shift+tab", "left":
			// Move to previous tab
			m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
			return m, nil
		}
	}

	if m.activeTab < len(m.tabs) {
		// Process the update for the active tab
		tabModel, cmd := m.tabs[m.activeTab].Update(msg)

		// Update the tab in the slice if it implements TabModel
		if updatedTab, ok := tabModel.(TabModel); ok {
			m.tabs[m.activeTab] = updatedTab
		}
		return m, cmd
	}

	return m, nil
}

func (m TabsModel) View() string {
	var header strings.Builder
	for i, tab := range m.tabs {
		if i == m.activeTab {
			header.WriteString(lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				Padding(0, 1).
				Render("[ " + tab.TabTitle() + " ]"))
		} else {
			header.WriteString(lipgloss.NewStyle().
				Padding(0, 1).
				Render("  " + tab.TabTitle() + "  "))
		}
	}

	content := m.tabs[m.activeTab].View()
	helpText := "← → Navigate tabs • j/k Up/Down • q Quit"
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(helpText)

	return lipgloss.JoinVertical(lipgloss.Left, header.String(), content, help)
}

func (m TabsModel) Init() tea.Cmd {
	return nil
}
