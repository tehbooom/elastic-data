package run

import (
	"context"
	"fmt"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/log"
	ProgramContext "github.com/tehbooom/elastic-data/ui/context"
)

const (
	StopedMsg  string = "Waiting to start"
	StartedMsg string = "Running"
)

// TabModel represents the integrations tab
type TabModel struct {
	width          int
	height         int
	programContext *ProgramContext.ProgramContext
	saveController *ProgramContext.SaveController
	// the key is integrationName:datasetName
	integrations map[string]*IntegrationStats
	table        *table.Table
	status       string
	running      bool
	error        error
	// list of installed integrations
	installedIntegrations []string
	generators            map[string]*DataGenerator
	mu                    sync.RWMutex
	mainCtx               context.Context
	mainCancel            context.CancelFunc
	wg                    sync.WaitGroup
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
func NewTabModel(programContext *ProgramContext.ProgramContext, saveController *ProgramContext.SaveController) *TabModel {
	ctx, cancel := context.WithCancel(context.Background())
	model := &TabModel{
		programContext:        programContext,
		saveController:        saveController,
		integrations:          make(map[string]*IntegrationStats),
		status:                StopedMsg,
		installedIntegrations: []string{},
		mainCtx:               ctx,
		mainCancel:            cancel,
		generators:            make(map[string]*DataGenerator),
		running:               false,
	}
	model.RefreshIntegrations()

	return model
}

// TabTitle returns the title of the tab
func (m *TabModel) TabTitle() string {
	return "Run"
}

// Init initializes the tab
func (m *TabModel) Init() tea.Cmd {
	log.Debug("TabModel.Init() - starting ticker")
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

type RunTabModel struct {
	TabModel *TabModel
}

func NewRunTabModel(context *ProgramContext.ProgramContext, saveController *ProgramContext.SaveController) *RunTabModel {
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
	log.Debug(fmt.Sprintf("RunTabModel.Update received: %T", msg))
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
