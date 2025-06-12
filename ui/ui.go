package ui

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	es "github.com/elastic/go-elasticsearch/v9"
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
	kb "github.com/tehbooom/go-kibana"
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
	programContext *ProgramContext.ProgramContext
	screen         Screen
	loading        LoadingModel
	tabs           tabs.TabsModel
	saveController *ProgramContext.SaveController
	error          *errors.ErrorOverlay
}

type ConfigLoadedMsg struct {
	Config     *config.Config
	ConfigPath string
	ESClient   *es.TypedClient
	KBClient   *kb.Client
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
		tabs:           tabs.NewTabsModel(initTabs, programContext),
	}
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		cfg, cfgPath, err := config.LoadConfig()
		if err != nil {
			log.Debug(err)
			return errors.ShowErrorMsg{Message: fmt.Sprintf("Error loading config: %v", err), Fatal: true}
		}

		esClient, err := elasticsearch.SetClient(cfg.Connection)
		if err != nil {
			log.Debug(err)
			return errors.ShowErrorMsg{Message: fmt.Sprintf("Error setting up Elasticsearch client: %v", err), Fatal: true}
		}

		kbClient, err := kibana.SetClient(cfg.Connection)
		if err != nil {
			log.Debug(err)
			return errors.ShowErrorMsg{Message: fmt.Sprintf("Error setting up Kibana client: %v", err), Fatal: true}
		}

		return ConfigLoadedMsg{
			Config:     cfg,
			ConfigPath: cfgPath,
			ESClient:   esClient,
			KBClient:   kbClient,
		}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.QuitMsg:
		return m, tea.Quit
	case ConfigLoadedMsg:
		m.programContext.ConfigPath = msg.ConfigPath
		m.programContext.Config = msg.Config

		if len(msg.Config.Integrations) > 0 {
			for integration, integrationData := range msg.Config.Integrations {
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
						Events:                configDataset.Events,
					}
					datasetMap[datasetName] = datasetConfig
				}
			}
		}

		ctx := context.Background()
		m.programContext.ESClient = &elasticsearch.Config{
			Client:    msg.ESClient,
			Ctx:       ctx,
			Connected: false,
		}
		m.programContext.KBClient = &kibana.Config{
			Client:    msg.KBClient,
			Ctx:       context.Background(),
			Connected: false,
		}

		return m, m.loading.Init()

	case errors.ShowErrorMsg:
		if msg.Fatal {
			log.Debug("Fatal error: %s", msg.Message)

			m.error = errors.NewErrorOverlay(msg.Message, true)
			return m, tea.Batch(
				errors.ErrorTimeout(),
				tea.Tick(5*time.Second, func(time.Time) tea.Msg {
					return tea.QuitMsg{}
				}),
			)
		} else {
			m.error = errors.NewErrorOverlay(msg.Message, false)
			return m, errors.ErrorTimeout()
		}
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

func (m Model) HasFatalError() (bool, string) {
	if m.error != nil && m.error.Fatal {
		return true, m.error.Message
	}
	return false, ""
}
