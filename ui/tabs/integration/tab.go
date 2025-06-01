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

	"github.com/tehbooom/elastic-data/ui/context"
	"github.com/tehbooom/elastic-data/ui/style"
)

const (
	StateSelectingIntegration = iota
	StateSelectingDatasets
	StateConfiguringDataset
)

type TabModel struct {
	width              int
	height             int
	context            *context.ProgramContext
	saveController     *context.SaveController
	state              int
	integrationList    list.Model
	datasetsList       list.Model
	thresholdInput     textinput.Model
	unitInput          textinput.Model
	viewport           viewport.Model
	currentIntegration string
	columnsPerRow      int
	selectedIndex      int
	scrollOffset       int
	visibleRows        int
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

func NewTabModel(context *context.ProgramContext, saveController *context.SaveController) *TabModel {
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

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Available Integrations"
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)
	l.SetShowPagination(true)

	datasetDelegate := NewCompactDelegate()
	datasetsList := list.New([]list.Item{}, datasetDelegate, 0, 0)
	datasetsList.SetShowHelp(false)
	datasetsList.SetShowStatusBar(false)
	datasetsList.SetShowPagination(false)
	datasetsList.Styles.Title = style.TitleStyle
	return &TabModel{
		integrationList: l,
		context:         context,
		viewport:        vp,
		datasetsList:    datasetsList,
		thresholdInput:  thInput,
		unitInput:       uInput,
		saveController:  saveController,
		state:           StateSelectingIntegration,
		scrollOffset:    0,
		visibleRows:     1,
	}
}

func (m TabModel) TabTitle() string {
	return "Integrations"
}

func (m TabModel) Init() tea.Cmd {
	return nil
}

func (m *TabModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.integrationList.SetSize(width, height)
	m.datasetsList.SetSize(width, height)
}

func (m *TabModel) IsInConfigurationState() bool {
	return m.state == StateConfiguringDataset
}
