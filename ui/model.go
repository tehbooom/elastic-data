package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tehbooom/elastic-data/internal/config"
)

// Screen represents the different screens in the application
type Screen int

const (
	LoadingScreen Screen = iota
	SelectionScreen
	DatasetScreen
	MetricsScreen
	ConnectionScreen
	ConfirmScreen
)

type Model struct {
	width                int
	height               int
	help                 help.Model
	configFilePath       string
	keys                 keyMap
	screen               Screen
	loading              LoadingModel
	selection            SelectionModel
	dataset              DatasetModel
	metrics              MetricsModel
	connection           ConnectionModel
	confirm              ConfirmModel
	selectedIntegrations []string
	selectedDatasets     map[string][]string
	selectedMetrics      map[string]string
	connectionDetails    ConnectionDetails
}

type ConnectionDetails struct {
	URL      string
	Username string
	Password string
}

func NewModel(cfgFile string) Model {
	h := help.New()
	h.ShowAll = false

	return Model{
		help:             h,
		configFilePath:   cfgFile,
		keys:             defaultKeyMap(),
		screen:           LoadingScreen,
		loading:          NewLoadingModel(),
		selection:        NewSelectionModel(),
		dataset:          NewDatasetModel(),
		metrics:          NewMetricsModel(),
		connection:       NewConnectionModel(),
		confirm:          NewConfirmModel(),
		selectedDatasets: make(map[string][]string),
		selectedMetrics:  make(map[string]string),
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	cfg, err := config.LoadConfig(m.configFilePath)
	if err != nil {
		// Handle error - maybe show an error screen
		// For now, we'll just log to console
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
	}

	if cfg != nil {
		// urls, username, password := cfg.GetConnectionDetails()
		// m.connection = m.connection.WithConnectionDetails(ConnectionDetails{
		// 	URLs:     urls,
		// 	Username: username,
		// 	Password: password,
		// })
		//
		// m.selection = m.selection.WithItems(cfg.GetIntegrations())

		// m.selectedDatasets = cfg.GetSelectedDatasets()
		// m.selectedMetrics = cfg.GetSelectedMetrics()
	}

	return m.loading.Init()
}

// Update handles all incoming events and updates the model accordingly
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width

		// Update size for all submodels
		m.loading.SetSize(msg.Width, msg.Height)
		m.selection.SetSize(msg.Width, msg.Height)
		m.dataset.SetSize(msg.Width, msg.Height)
		m.metrics.SetSize(msg.Width, msg.Height)
		m.connection.SetSize(msg.Width, msg.Height)
		m.confirm.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		// Global keyboard shortcuts
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}
	}

	// Screen-specific updates
	switch m.screen {
	case LoadingScreen:
		loadingModel, cmd := m.loading.Update(msg)
		m.loading = loadingModel.(LoadingModel)

		// Check if loading is complete
		if m.loading.IsComplete() {
			m.screen = SelectionScreen
			cmds = append(cmds, m.selection.Init())
		}

		cmds = append(cmds, cmd)

	case SelectionScreen:
		selectionModel, cmd := m.selection.Update(msg)
		m.selection = selectionModel.(SelectionModel)

		// Check if selection is complete
		if m.selection.IsComplete() {
			m.selectedIntegrations = m.selection.GetSelectedItems()
			m.dataset.SetAvailableItems(m.selectedIntegrations)
			m.screen = DatasetScreen
			cmds = append(cmds, m.dataset.Init())
		}

		cmds = append(cmds, cmd)

	case DatasetScreen:
		datasetModel, cmd := m.dataset.Update(msg)
		m.dataset = datasetModel.(DatasetModel)

		// Check if dataset selection is complete
		if m.dataset.IsComplete() {
			m.selectedDatasets = m.dataset.GetSelectedDatasets()
			m.metrics.SetSelectedDatasets(m.selectedDatasets)
			m.screen = MetricsScreen
			cmds = append(cmds, m.metrics.Init())
		}

		cmds = append(cmds, cmd)

	case MetricsScreen:
		metricsModel, cmd := m.metrics.Update(msg)
		m.metrics = metricsModel.(MetricsModel)

		// Check if metrics selection is complete
		if m.metrics.IsComplete() {
			m.selectedMetrics = m.metrics.GetSelectedMetrics()
			m.screen = ConnectionScreen
			cmds = append(cmds, m.connection.Init())
		}

		cmds = append(cmds, cmd)

	case ConnectionScreen:
		connectionModel, cmd := m.connection.Update(msg)
		m.connection = connectionModel.(ConnectionModel)

		// Check if connection details are complete
		if m.connection.IsComplete() {
			m.connectionDetails = m.connection.GetConnectionDetails()

			// Test connection
			if m.connection.IsConnectionValid() {
				m.screen = ConfirmScreen
				cmds = append(cmds, m.confirm.Init())
			}
		}

		cmds = append(cmds, cmd)

	case ConfirmScreen:
		confirmModel, cmd := m.confirm.Update(msg)
		m.confirm = confirmModel.(ConfirmModel)

		// Check if confirmed to start
		if m.confirm.IsConfirmed() {
			// Start the process
			// This could be a message that's handled elsewhere
			return m, tea.Quit
		}

		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the current UI
func (m Model) View() string {
	var content string

	switch m.screen {
	case LoadingScreen:
		content = m.loading.View()
	case SelectionScreen:
		content = m.selection.View()
	case DatasetScreen:
		content = m.dataset.View()
	case MetricsScreen:
		content = m.metrics.View()
	case ConnectionScreen:
		content = m.connection.View()
	case ConfirmScreen:
		content = m.confirm.View()
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		content,
	)
}

// keyMap defines the keybindings for the application
type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Select key.Binding
	Help   key.Binding
	Quit   key.Binding
}

// defaultKeyMap returns the default keybindings
func defaultKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "move left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "move right"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter/space", "select"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q/ctrl+c", "quit"),
		),
	}
}
