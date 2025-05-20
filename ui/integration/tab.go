package integration

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	// "github.com/tehbooom/elastic-data/internal/integrations"
	"github.com/tehbooom/elastic-data/ui/state"
)

const (
	StateSelectingIntegration = iota
	StateSelectingDatasets
	StateConfiguringDataset
)

// TabModel represents the integrations tab
type TabModel struct {
	width              int
	height             int
	integrationList    list.Model
	appState           *state.AppState
	state              int
	datasetsList       list.Model
	thresholdInput     textinput.Model
	unitInput          textinput.Model
	currentIntegration string
	viewport           viewport.Model
	saveController     *state.SaveController
}

func ValidateUnit(input string) error {
	lowerInput := strings.ToLower(input)

	if lowerInput == "eps" || lowerInput == "bytes" {
		return nil
	}

	return errors.New("Unit must be 'eps' or 'bytes'")
}

func ValidateThreshold(input string) error {
	_, err := strconv.Atoi(input)
	if err != nil {
		return fmt.Errorf("Threshold value is not a number")
	}

	return nil
}

// NewTabModel creates a new integrations tab model
func NewTabModel(state *state.AppState, saveController *state.SaveController) *TabModel {
	thInput := textinput.New()
	thInput.Placeholder = "Enter threshold value"
	thInput.CharLimit = 10
	// TODO: Verify only ints are entered
	//thInput.Validate()

	uInput := textinput.New()
	uInput.Placeholder = "Unit (eps or bytes)"
	uInput.SetSuggestions([]string{"eps", "bytes"})
	uInput.ShowSuggestions = true
	uInput.CharLimit = 5
	uInput.Validate = ValidateUnit

	vp := viewport.New(0, 0)
	delegate := NewCompactDelegate()

	// Setup integration list
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Available Integrations"
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)
	l.SetShowPagination(true)

	// Setup datasets list
	datasetsList := list.New([]list.Item{}, delegate, 0, 0)
	datasetsList.SetShowHelp(false)
	datasetsList.SetShowStatusBar(false)
	datasetsList.SetShowPagination(false)

	return &TabModel{
		integrationList: l,
		appState:        state,
		viewport:        vp,
		datasetsList:    datasetsList,
		thresholdInput:  thInput,
		unitInput:       uInput,
		saveController:  saveController,
		state:           StateSelectingIntegration,
	}
}

// TabTitle returns the title of the tab
func (m TabModel) TabTitle() string {
	return "Integrations"
}

// Init initializes the tab
func (m TabModel) Init() tea.Cmd {
	return nil
}

// SetSize sets the size of the tab
func (m *TabModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.integrationList.SetSize(width, height)
	m.datasetsList.SetSize(width, height)
}

// IsInConfigurationState returns true if the tab is in configuration state
func (m *TabModel) IsInConfigurationState() bool {
	return m.state == StateConfiguringDataset
}
