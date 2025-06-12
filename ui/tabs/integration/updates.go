package integration

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/tehbooom/elastic-data/ui/context"
	"github.com/tehbooom/elastic-data/ui/errors"
)

func (m *TabModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmd := m.handleGlobalKeys(msg)
	if cmd != nil {
		return m, cmd
	}

	switch m.state {
	case StateSelectingIntegration:
		return m.updateIntegrationSelection(msg)
	case StateSelectingDatasets:
		return m.updateDatasetSelection(msg)
	case StateConfiguringDataset:
		return m.updateDatasetConfiguration(msg)
	}

	return m, nil
}

func (m *TabModel) handleGlobalKeys(msg tea.Msg) tea.Cmd {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "tab", "right", "shift+tab", "left":
			if m.state != StateConfiguringDataset {
				return nil
			}

			if keyMsg.String() == "shift+tab" && m.state == StateConfiguringDataset {
				if m.thresholdInput.Focused() {
					m.thresholdInput.Blur()
					m.unitInput.Focus()
				} else {
					m.unitInput.Blur()
					m.thresholdInput.Focus()
				}
				return nil
			}
		}
	}
	return nil
}

func (m *TabModel) updateIntegrationSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.searchMode {
			switch msg.String() {
			case "enter":
				m.searchMode = false
				m.selectedIndex = 0
				return m, nil
			case "esc":
				m.searchMode = false
				m.searchQuery = ""
				m.filteredItems = nil
				m.selectedIndex = 0
				return m, nil
			case "backspace":
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
					m.filteredItems = m.filterItems(m.integrationList.Items(), m.searchQuery)
					m.selectedIndex = 0
				}
				return m, nil
			default:
				if len(msg.String()) == 1 && msg.String() != " " || msg.String() == "space" {
					char := msg.String()
					if char == "space" {
						char = " "
					}
					m.searchQuery += char
					m.filteredItems = m.filterItems(m.integrationList.Items(), m.searchQuery)
					m.selectedIndex = 0
				}
				return m, nil
			}
		} else {
			items := m.integrationList.Items()
			if m.searchMode || m.filteredItems != nil {
				items = m.filteredItems
			}

			totalItems := len(items)

			switch msg.String() {
			case "/":
				m.searchMode = true
				m.onlySelected = false
				m.searchQuery = ""
				m.filteredItems = m.integrationList.Items()
				return m, nil
			}

			if m.handleGridNavigation(msg, totalItems) {
				return m, nil
			}

			switch msg.String() {
			case " ":
				if m.selectedIndex < totalItems {
					var item *IntegrationItem
					var ok bool
					if m.filteredItems != nil {
						item, ok = m.filteredItems[m.selectedIndex].(*IntegrationItem)
					} else {
						item, ok = m.integrationList.Items()[m.selectedIndex].(*IntegrationItem)
					}
					if !ok {
						return m, nil
					}
					item.Selected = !item.Selected
					m.context.SetIntegrationSelected(item.Name, item.Selected)
					m.saveController.MarkDirty()

					originalItems := m.integrationList.Items()
					for i, origItem := range originalItems {
						if origIntegration, ok := origItem.(*IntegrationItem); ok && origIntegration.Name == item.Name {
							originalItems[i] = item
							break
						}
					}
					m.integrationList.SetItems(originalItems)

					if m.searchMode && m.filteredItems != nil {
						m.filteredItems[m.selectedIndex] = item
					}
				}
				return m, nil

			case "e":
				m.onlySelected = !m.onlySelected
				if m.onlySelected {
					m.filteredItems = m.viewSelected(m.integrationList.Items())
				} else {
					m.filteredItems = nil
				}
				m.selectedIndex = 0
				return m, nil

			case "enter":
				if m.selectedIndex < totalItems {
					var item *IntegrationItem
					var ok bool
					if m.filteredItems != nil {
						item, ok = m.filteredItems[m.selectedIndex].(*IntegrationItem)
					} else {
						item, ok = m.integrationList.Items()[m.selectedIndex].(*IntegrationItem)
					}
					if !ok {
						return m, nil
					}
					if !item.Selected {
						item.Selected = true
						m.context.SetIntegrationSelected(item.Name, item.Selected)
						m.saveController.MarkDirty()

						originalItems := m.integrationList.Items()
						for i, origItem := range originalItems {
							if origIntegration, ok := origItem.(*IntegrationItem); ok && origIntegration.Name == item.Name {
								originalItems[i] = item
								break
							}
						}
						m.integrationList.SetItems(originalItems)
					}
					m.currentIntegration = item.Name
					err := m.loadDatasetsForIntegration(item.Name)
					if err != nil {
						log.Debug(err)
						return m, func() tea.Msg {
							return errors.ShowErrorMsg{Message: fmt.Sprintf("Error: %v", err)}
						}
					}
					m.state = StateSelectingDatasets

					m.selectedIndex = 0
					m.scrollOffset = 0
					return m, nil
				}
			case "esc", "q":
				m.searchMode = false
				m.onlySelected = false
				m.searchQuery = ""
				m.filteredItems = nil
				m.selectedIndex = 0
				m.readmeRendered = false
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.integrationList, cmd = m.integrationList.Update(msg)
	return m, cmd
}

func (m *TabModel) updateDatasetSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.focusedDatasetComponent == FocusDatasetList {
				log.Debug(fmt.Sprintf("index is %d", m.datasetsList.Index()))
				if m.datasetsList.Index() < len(m.datasetsList.Items())-1 {
					var cmd tea.Cmd
					m.datasetsList, cmd = m.datasetsList.Update(msg)
					return m, cmd
				} else {
					m.focusedDatasetComponent = FocusViewport
					m.viewport.Style = lipgloss.NewStyle().
						Border(lipgloss.RoundedBorder()).
						BorderForeground(lipgloss.Color("62"))
					m.lastListIndex = m.datasetsList.Index()
					if len(m.datasetsList.Items()) > 1 {
						m.datasetsList.Select(-1)
					}
					return m, nil
				}
			} else {
				var cmd tea.Cmd
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd
			}
		case "k", "up":
			if m.focusedDatasetComponent == FocusViewport {
				if m.viewport.AtTop() {
					m.focusedDatasetComponent = FocusDatasetList
					m.viewport.Style = lipgloss.NewStyle().
						Border(lipgloss.RoundedBorder()).
						BorderForeground(lipgloss.Color("240"))
					m.datasetsList.Select(m.lastListIndex)
					return m, nil
				} else {
					var cmd tea.Cmd
					m.viewport, cmd = m.viewport.Update(msg)
					return m, cmd
				}
			} else {
				var cmd tea.Cmd
				m.datasetsList, cmd = m.datasetsList.Update(msg)
				return m, cmd
			}
		case " ":
			if m.focusedDatasetComponent == FocusDatasetList {
				item, ok := m.datasetsList.SelectedItem().(DatasetItem)
				if !ok {
					return m, nil
				}
				item.Selected = !item.Selected
				items := m.datasetsList.Items()
				items[m.datasetsList.Index()] = item
				m.datasetsList.SetItems(items)
				m.updateDatasetConfigs()
			}
			return m, nil

		case "enter":
			if m.focusedDatasetComponent == FocusDatasetList {
				item, ok := m.datasetsList.SelectedItem().(DatasetItem)
				if !ok {
					return m, nil
				}

				if !item.Selected {
					items := m.datasetsList.Items()
					items[m.datasetsList.Index()] = item
					m.datasetsList.SetItems(items)
					m.updateDatasetConfigs()
				}

				m.thresholdInput.SetValue(strconv.Itoa(item.Threshold))
				m.thresholdInput.Focus()
				m.unitInput.SetValue(item.Unit)
				m.preserveInput.SetValue(strconv.FormatBool(item.PreserveEventOriginal))
				m.state = StateConfiguringDataset
			}
			return m, nil

		case "esc", "q":
			m.readmeRendered = false
			m.focusedDatasetComponent = FocusDatasetList
			m.viewport.AtTop()
			m.datasetsList.Select(0)
			m.state = StateSelectingIntegration
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.datasetsList, cmd = m.datasetsList.Update(msg)
	return m, cmd
}

func (m *TabModel) updateDatasetConfiguration(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			if m.thresholdInput.Focused() {
				m.thresholdInput.Blur()
				m.unitInput.Focus()
			} else if m.unitInput.Focused() {
				m.unitInput.Blur()
				m.preserveInput.Focus()
			} else if m.preserveInput.Focused() {
				m.preserveInput.Blur()
				m.thresholdInput.Focus()
			} else {
				m.thresholdInput.Focus()
			}
			return m, nil

		case "enter":
			idx := m.datasetsList.Index()
			items := m.datasetsList.Items()
			item, ok := items[idx].(DatasetItem)
			if !ok {
				return m, nil
			}

			threshold, _ := strconv.Atoi(m.thresholdInput.Value())
			unit := m.unitInput.Value()
			preserve, err := strconv.ParseBool(m.preserveInput.Value())
			if err != nil {
				preserve = false
			}

			if threshold <= 0 {
				return m, func() tea.Msg {
					return errors.ShowErrorMsg{Message: "threshold must be greater than 0"}
				}
			}

			item.Threshold = threshold
			item.Unit = unit
			items[idx] = item
			m.datasetsList.SetItems(items)

			datasetMap, exists := m.context.DatasetConfigs[m.currentIntegration]
			if !exists {
				datasetMap = make(map[string]context.DatasetConfig)
				m.context.DatasetConfigs[m.currentIntegration] = datasetMap
			}

			datasetMap[item.Name] = context.DatasetConfig{
				Name:                  item.Name,
				Selected:              item.Selected,
				Threshold:             threshold,
				Unit:                  unit,
				PreserveEventOriginal: preserve,
				Events:                item.Events,
			}

			if !item.Selected {
				item.Selected = true
			}

			m.saveController.MarkDirty()

			m.state = StateSelectingDatasets
			return m, nil

		case "esc", "q":
			m.state = StateSelectingDatasets
			return m, nil
		}
	}

	var cmd tea.Cmd
	if m.thresholdInput.Focused() {
		m.thresholdInput, cmd = m.thresholdInput.Update(msg)
	} else if m.unitInput.Focused() {
		m.unitInput, cmd = m.unitInput.Update(msg)
	} else {
		m.preserveInput, cmd = m.preserveInput.Update(msg)
	}
	return m, cmd
}

func (m *TabModel) handleGridNavigation(msg tea.KeyMsg, totalItems int) bool {
	if totalItems == 0 {
		return false
	}

	columns := m.calculateColumns()
	currentRow := m.selectedIndex / columns
	currentCol := m.selectedIndex % columns
	totalRows := (totalItems + columns - 1) / columns
	oldIndex := m.selectedIndex

	switch msg.String() {
	case "k", "up":
		if currentRow > 0 {
			newIndex := (currentRow-1)*columns + currentCol
			if newIndex >= 0 {
				m.selectedIndex = newIndex
			}
		}
	case "j", "down":
		if currentRow < totalRows-1 {
			newIndex := (currentRow+1)*columns + currentCol
			if newIndex < totalItems {
				m.selectedIndex = newIndex
			} else {
				lastRowStart := (totalRows - 1) * columns
				if lastRowStart+currentCol < totalItems {
					m.selectedIndex = lastRowStart + currentCol
				} else {
					m.selectedIndex = totalItems - 1
				}
			}
		}
	case "h", "left":
		if currentCol > 0 {
			m.selectedIndex--
		} else if currentRow > 0 {
			prevRowStart := (currentRow - 1) * columns
			prevRowEnd := min(prevRowStart+columns-1, totalItems-1)
			m.selectedIndex = prevRowEnd
		}
	case "l", "right":
		if currentCol < columns-1 && m.selectedIndex < totalItems-1 {
			nextIndex := m.selectedIndex + 1
			nextRow := nextIndex / columns
			if nextRow == currentRow {
				m.selectedIndex = nextIndex
			}
		} else if currentRow < totalRows-1 {
			m.selectedIndex = (currentRow + 1) * columns
		}
	case "pgup", "ctrl+u":
		newRow := currentRow - m.visibleRows
		if newRow < 0 {
			newRow = 0
		}
		m.selectedIndex = newRow*columns + currentCol
		if m.selectedIndex >= totalItems {
			m.selectedIndex = totalItems - 1
		}
	case "pgdown", "ctrl+d":
		newRow := currentRow + m.visibleRows
		if newRow >= totalRows {
			newRow = totalRows - 1
		}
		newIndex := newRow*columns + currentCol
		if newIndex >= totalItems {
			m.selectedIndex = totalItems - 1
		} else {
			m.selectedIndex = newIndex
		}
	case "home", "g":
		m.selectedIndex = 0
	case "end", "G":
		m.selectedIndex = totalItems - 1
	default:
		return false
	}

	if m.selectedIndex != oldIndex {
		m.ensureSelectionVisible(totalItems)
	}

	return true
}
