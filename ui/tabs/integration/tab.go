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
	"github.com/charmbracelet/lipgloss"

	"github.com/tehbooom/elastic-data/ui/context"
	"github.com/tehbooom/elastic-data/ui/style"
)

const (
	StateSelectingIntegration = iota
	StateSelectingDatasets
	StateConfiguringDataset

	FocusDatasetList = iota
	FocusViewport
)

type TabModel struct {
	width                   int
	height                  int
	context                 *context.ProgramContext
	saveController          *context.SaveController
	state                   int
	integrationList         list.Model
	datasetsList            list.Model
	thresholdInput          textinput.Model
	unitInput               textinput.Model
	preserveInput           textinput.Model
	viewport                viewport.Model
	currentIntegration      string
	selectedIndex           int
	scrollOffset            int
	visibleRows             int
	readmeRendered          bool
	focusedDatasetComponent int
	lastListIndex           int
	searchMode              bool
	searchQuery             string
	filteredItems           []list.Item
	onlySelected            bool
}

func ValidateUnit(input string) error {
	lowerInput := strings.ToLower(input)

	if lowerInput == "eps" || lowerInput == "bytes" {
		return nil
	}

	return errors.New("unit must be 'eps' or 'bytes'")
}

func ValidateThreshold(input string) error {
	value, err := strconv.Atoi(input)
	if err != nil {
		return fmt.Errorf("threshold value is not a number")
	}

	if value < 1 {
		return fmt.Errorf("threshold value must be greater than or equal to 1")
	}

	return nil
}

func NewTabModel(context *context.ProgramContext, saveController *context.SaveController) *TabModel {
	thInput := textinput.New()
	thInput.Placeholder = "Enter threshold value"
	thInput.CharLimit = 10
	thInput.Validate = ValidateThreshold

	uInput := textinput.New()
	uInput.Placeholder = "Unit (eps or bytes)"
	uInput.SetSuggestions([]string{"eps", "bytes"})
	uInput.ShowSuggestions = true
	uInput.CharLimit = 5
	uInput.Validate = ValidateUnit

	pInput := textinput.New()
	pInput.Placeholder = "false"
	pInput.SetSuggestions([]string{"true", "false"})
	pInput.ShowSuggestions = true

	delegate := NewCompactDelegate()

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Available Integrations"
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.SetShowStatusBar(false)
	l.SetShowPagination(true)

	datasetDelegate := NewCompactDelegate()
	datasetsList := list.New([]list.Item{}, datasetDelegate, 0, 0)
	datasetsList.SetShowHelp(false)
	datasetsList.SetShowStatusBar(false)
	datasetsList.SetShowPagination(false)
	datasetsList.Styles.Title = style.TitleStyle

	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	return &TabModel{
		integrationList:         l,
		context:                 context,
		datasetsList:            datasetsList,
		thresholdInput:          thInput,
		unitInput:               uInput,
		preserveInput:           pInput,
		saveController:          saveController,
		state:                   StateSelectingIntegration,
		scrollOffset:            0,
		visibleRows:             1,
		viewport:                vp,
		readmeRendered:          false,
		focusedDatasetComponent: FocusDatasetList,
		onlySelected:            false,
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

	helpHeight := 2
	paddingHeight := 4
	viewportHeight := height/2 - 5

	availableHeight := height - viewportHeight - helpHeight - paddingHeight
	listHeight := max(availableHeight, 5)
	m.datasetsList.SetSize(width, listHeight)
}

func (m *TabModel) IsInConfigurationState() bool {
	return m.state == StateConfiguringDataset
}
