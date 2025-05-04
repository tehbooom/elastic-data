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
	"github.com/tehbooom/elastic-data/ui/state"
)

// Screen represents the different screens in the application
type Screen int

const (
	LoadingScreen Screen = iota
	TabsScreen
)

type Model struct {
	width          int
	height         int
	help           help.Model
	keys           keyMap
	state          *state.AppState
	screen         Screen
	loading        LoadingModel
	tabs           TabsModel
	saveController *state.SaveController
}

type TabModel interface {
	tea.Model
	SetSize(width, height int)
	TabTitle() string
}

func NewModel() Model {
	appState := state.NewAppState()
	saveController := state.NewSaveController(appState)
	h := help.New()
	h.ShowAll = false

	integrationsTab := NewIntegrationsTabModel(appState, saveController)

	tabs := []TabModel{integrationsTab}
	return Model{
		help:           h,
		state:          appState,
		saveController: saveController,
		screen:         LoadingScreen,
		loading:        NewLoadingModel(),
		tabs:           NewTabsModel(tabs),
	}
}

func (m Model) Init() tea.Cmd {
	cfg, cfgPath, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
	}

	if cfg != nil {
		m.state.ConfigPath = cfgPath
		m.state.Config = cfg
		if len(cfg.Integrations) > 0 {
			for integration, integrationData := range cfg.Integrations {
				if integrationData.Enabled {
					m.state.SelectedIntegrations[integration] = true
				}

				datasetMap, exists := m.state.DatasetConfigs[integration]
				if !exists {
					datasetMap = make(map[string]state.DatasetConfig)
					m.state.DatasetConfigs[integration] = datasetMap
				}

				for datasetName, configDataset := range integrationData.Datasets {
					datasetConfig := state.DatasetConfig{
						Name:      datasetName,
						Selected:  configDataset.Enabled,
						Unit:      configDataset.Unit,
						Threshold: configDataset.Threshold,
					}
					datasetMap[datasetName] = datasetConfig
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
		switch msg.String() {
		case "ctrl+c", "q":
			m.saveController.SaveNow()
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
			if m.state.SelectedIntegrations == nil {
				m.state.SelectedIntegrations = make(map[string]bool)
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
