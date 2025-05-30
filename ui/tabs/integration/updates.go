package integration

import (
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tehbooom/elastic-data/ui/context"
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
		totalItems := len(m.integrationList.Items())
		if m.handleGridNavigation(msg, totalItems) {
			return m, nil
		}
		switch msg.String() {
		case " ":
			if m.selectedIndex < totalItems {
				item, ok := m.integrationList.Items()[m.selectedIndex].(*IntegrationItem)
				if !ok {
					return m, nil
				}
				item.Selected = !item.Selected
				m.context.SetIntegrationSelected(item.Name, item.Selected)
				m.saveController.MarkDirty()
				items := m.integrationList.Items()
				items[m.selectedIndex] = item
				m.integrationList.SetItems(items)
				// items := m.integrationList.Items()
				// items[m.integrationList.Index()] = item
				// m.integrationList.SetItems(items)
			}
			return m, nil

		case "enter":
			if m.selectedIndex < totalItems {
				item, ok := m.integrationList.Items()[m.selectedIndex].(*IntegrationItem)
				if !ok {
					return m, nil
				}

				if !item.Selected {
					item.Selected = true
					m.context.SetIntegrationSelected(item.Name, item.Selected)
					m.saveController.MarkDirty()
					items := m.integrationList.Items()
					items[m.selectedIndex] = item
					m.integrationList.SetItems(items)
					// items := m.integrationList.Items()
					// items[m.integrationList.Index()] = item
					// m.integrationList.SetItems(items)
				}

				m.currentIntegration = item.Name
				m.loadDatasetsForIntegration(item.Name)
				m.state = StateSelectingDatasets
				m.selectedIndex = 0
				m.scrollOffset = 0
				return m, nil
			}

		case "esc", "q":
			return m, nil
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
		case " ":
			item, ok := m.datasetsList.SelectedItem().(DatasetItem)
			if !ok {
				return m, nil
			}
			item.Selected = !item.Selected
			items := m.datasetsList.Items()
			items[m.datasetsList.Index()] = item
			m.datasetsList.SetItems(items)
			m.updateDatasetConfigs()
			return m, nil

		case "enter":
			item, ok := m.datasetsList.SelectedItem().(DatasetItem)
			if !ok {
				return m, nil
			}

			if !item.Selected {
				item.Selected = true
				items := m.datasetsList.Items()
				items[m.datasetsList.Index()] = item
				m.datasetsList.SetItems(items)
				m.updateDatasetConfigs()
			}

			m.thresholdInput.SetValue(strconv.Itoa(item.Threshold))
			m.thresholdInput.Focus()
			m.unitInput.SetValue(item.Unit)
			m.state = StateConfiguringDataset
			return m, nil

		case "left", "esc", "q":
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
			} else {
				m.unitInput.Blur()
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
				Name:      item.Name,
				Selected:  item.Selected,
				Threshold: threshold,
				Unit:      unit,
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
