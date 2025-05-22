package run

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/tehbooom/elastic-data/ui/state"
)

// TabModel represents the integrations tab
type TabModel struct {
	width          int
	height         int
	appState       *state.AppState
	saveController *state.SaveController
	integrations   map[string]IntegrationStats
	table          *table.Table
	status         string
	running        bool
	error          error
	//esClient
}

type IntegrationStats struct {
	Current   float64
	Peak      float64
	Unit      string
	LastValue float64
	Trend     string
}

type TabError struct {
	Message string
	Err     error
}

func errorCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return err
	}
}

func (e TabError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// NewTabModel creates a new run tab model
func NewTabModel(state *state.AppState, saveController *state.SaveController) *TabModel {
	model := &TabModel{
		appState:       state,
		saveController: saveController,
		integrations:   make(map[string]IntegrationStats),
		status:         "Waiting to start",
		running:        false,
	}
	model.RefreshIntegrations()

	return model
}

// TabTitle returns the title of the tab
func (m TabModel) TabTitle() string {
	return "Run"
}

// Init initializes the tab
func (m TabModel) Init() tea.Cmd {
	m.RefreshIntegrations()
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

// SetSize sets the size of the tab
func (m *TabModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}
