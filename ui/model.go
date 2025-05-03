package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tehbooom/elastic-data/internal/config"
	"github.com/tehbooom/elastic-data/internal/integrations"
)

// Screen represents the different screens in the application
type Screen int

const (
	LoadingScreen Screen = iota
	TabsScreen
)

type Model struct {
	width   int
	height  int
	help    help.Model
	keys    keyMap
	state   *AppState
	screen  Screen
	loading LoadingModel
	tabs    TabsModel
}

type TabModel interface {
	tea.Model
	SetSize(width, height int)
	TabTitle() string
}

func NewModel() Model {
	state := NewAppState()
	h := help.New()
	h.ShowAll = false

	integrationsTab := NewIntegrationsTabModel(state)

	tabs := []TabModel{integrationsTab}
	return Model{
		help:    h,
		state:   state,
		screen:  LoadingScreen,
		loading: NewLoadingModel(),
		tabs:    NewTabsModel(tabs),
	}
}

func (m Model) Init() tea.Cmd {
	cfg, cfgPath, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
	}

	if cfg != nil {
		m.state.configDir = cfgPath
		m.state.config = cfg
		if len(cfg.Integrations) > 0 {
			for _, integration := range cfg.Integrations {
				if integration.Enabled {
					m.state.selectedIntegrations[integration.Name] = true
				}
				integrationDatasets := m.state.datasetConfigs[integration.Name]
				for _, dataset := range integrationDatasets {
					dataset.Name
				}
			}
		}
	}

	return m.loading.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width

		// Set loading size
		m.loading.SetSize(msg.Width, msg.Height)

		// Always set tabs size regardless of current screen
		// This ensures the tabs are properly sized when we switch to them
		m.tabs.SetSize(msg.Width, m.height)

		// Return this update immediately to ensure size changes are applied
		return m, nil

	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}
	}

	switch m.screen {
	case LoadingScreen:
		loadingModel, cmd := m.loading.Update(msg)
		m.loading = loadingModel.(LoadingModel)
		if m.loading.IsComplete() {
			m.screen = TabsScreen
			configDir, _ := getConfigDir()
			repoDir := filepath.Join(configDir, "integrations")
			integrations, _ := integrations.GetIntegrations(repoDir)
			if m.state.selectedIntegrations == nil {
				m.state.selectedIntegrations = make(map[string]bool)
			}

			for _, tab := range m.tabs.tabs {
				if intTab, ok := tab.(*IntegrationsTabModel); ok {
					intTab.SetIntegrations(integrations)
					break
				}
			}
			m.tabs.SetSize(m.width, m.height)
			return m, nil
		}
		cmds = append(cmds, cmd)
	case TabsScreen:
		tabsModel, cmd := m.tabs.Update(msg)
		m.tabs = tabsModel.(TabsModel)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the current UI
func (m Model) View() string {
	var content string

	switch m.screen {
	case LoadingScreen:
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			m.loading.View(),
		)
	case TabsScreen:
		content = m.tabs.View()
	}

	return content
}

type keyMap struct {
	Quit key.Binding
}
