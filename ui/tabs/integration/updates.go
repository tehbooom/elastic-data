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
