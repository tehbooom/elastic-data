package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tehbooom/elastic-data/internal/config"
	"github.com/tehbooom/elastic-data/internal/elasticsearch"
	"github.com/tehbooom/elastic-data/internal/integrations"
	"github.com/tehbooom/elastic-data/internal/kibana"
	ProgramContext "github.com/tehbooom/elastic-data/ui/context"
	"github.com/tehbooom/elastic-data/ui/errors"
	"github.com/tehbooom/elastic-data/ui/style"
	"github.com/tehbooom/elastic-data/ui/tabs"
	"github.com/tehbooom/elastic-data/ui/tabs/integration"
	"github.com/tehbooom/elastic-data/ui/tabs/run"
)

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
	programContext *ProgramContext.ProgramContext
	screen         Screen
	loading        LoadingModel
	tabs           tabs.TabsModel
	saveController *ProgramContext.SaveController
	error          *errors.ErrorOverlay
}

func NewModel() Model {
	programContext := ProgramContext.NewProgramContext()
	saveController := ProgramContext.NewSaveController(programContext)
	h := help.New()
	h.ShowAll = false

	integrationsTab := integration.NewIntegrationsTabModel(programContext, saveController)
	runTab := run.NewRunTabModel(programContext, saveController)

	initTabs := []tabs.TabModel{integrationsTab, runTab}
	return Model{
		help:           h,
		programContext: programContext,
		saveController: saveController,
		screen:         LoadingScreen,
		loading:        NewLoadingModel(),
		tabs:           tabs.NewTabsModel(initTabs),
	}
}

func (m Model) Init() tea.Cmd {
	cfg, cfgPath, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
	}

	if cfg != nil {
		m.programContext.ConfigPath = cfgPath
		m.programContext.Config = cfg
		if len(cfg.Integrations) > 0 {
			for integration, integrationData := range cfg.Integrations {
				if integrationData.Enabled {
					m.programContext.SelectedIntegrations[integration] = true
				}

				datasetMap, exists := m.programContext.DatasetConfigs[integration]
				if !exists {
					datasetMap = make(map[string]ProgramContext.DatasetConfig)
					m.programContext.DatasetConfigs[integration] = datasetMap
				}

				for datasetName, configDataset := range integrationData.Datasets {
					datasetConfig := ProgramContext.DatasetConfig{
						Name:                  datasetName,
						Selected:              configDataset.Enabled,
						Unit:                  configDataset.Unit,
						Threshold:             configDataset.Threshold,
						PreserveEventOriginal: configDataset.PreserveEventOriginal,
					}
					datasetMap[datasetName] = datasetConfig
				}
			}
		}

		esClient, err := elasticsearch.SetClient(cfg.Connection)
		if err != nil {
			fmt.Println(err)
		}
		ctx := context.Background()
		m.programContext.ESClient = &elasticsearch.Config{
			Client:    esClient,
			Ctx:       ctx,
			Connected: false,
		}

		kbClient, err := kibana.SetClient(cfg.Connection)
		if err != nil {
			fmt.Println(err)
		}
		m.programContext.KBClient = &kibana.Config{
			Client:    kbClient,
			Ctx:       context.Background(),
			Connected: false,
		}
	}

	return m.loading.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case errors.ShowErrorMsg:
		m.error = errors.NewErrorOverlay(msg.Message)
		return m, errors.ErrorTimeout()
	case errors.ErrorTimeoutMsg:
		m.error = nil
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width

		m.loading.SetSize(msg.Width, msg.Height)

		m.tabs.SetSize(msg.Width, m.height)

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
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
			if m.programContext.SelectedIntegrations == nil {
				m.programContext.SelectedIntegrations = make(map[string]bool)
			}

			for _, tab := range m.tabs.Tabs {
				if intTab, ok := tab.(*integration.IntegrationsTabModel); ok {
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
		m.tabs = tabsModel.(tabs.TabsModel)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

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

	if m.error != nil {
		errorBox := style.ErrorStyle.Render("❌ " + m.error.Message)

		if m.width > 0 && m.height > 0 {
			errorOverlay := lipgloss.Place(
				m.width,
				m.height,
				lipgloss.Center,
				lipgloss.Center,
				errorBox,
			)
			return errorOverlay
		} else {
			return "\n\n" + errorBox
		}
	}

	return content
}

type keyMap struct {
	Quit key.Binding
}
