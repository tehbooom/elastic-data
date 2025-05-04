package integration

import (
	tea "github.com/charmbracelet/bubbletea"
	"strconv"
)

// Update handles user input and updates the model
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

// Handle global navigation keys
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

// Update for integration selection state
func (m *TabModel) updateIntegrationSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case " ": // Toggle selection
			item, ok := m.integrationList.SelectedItem().(*IntegrationItem)
			if !ok {
				return m, nil
			}
			item.Selected = !item.Selected

			m.appState.SetIntegrationSelected(item.Name, item.Selected)
			m.saveController.MarkDirty()

			items := m.integrationList.Items()
			items[m.integrationList.Index()] = item
			m.integrationList.SetItems(items)
			return m, nil

		case "enter":
			item, ok := m.integrationList.SelectedItem().(*IntegrationItem)
			if !ok {
				return m, nil
			}
			item.Selected = !item.Selected

			m.appState.SetIntegrationSelected(item.Name, item.Selected)
			m.saveController.MarkDirty()
			items := m.integrationList.Items()
			items[m.integrationList.Index()] = item
			m.integrationList.SetItems(items)

			m.currentIntegration = item.Name
			m.loadDatasetsForIntegration(item.Name)
			m.state = StateSelectingDatasets
			return m, nil

		case "esc", "q":
			return m, tea.Quit
		}
	}

	// Pass other messages to integration list
	var cmd tea.Cmd
	m.integrationList, cmd = m.integrationList.Update(msg)
	return m, cmd
}

// Update for dataset selection state
func (m *TabModel) updateDatasetSelection(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case " ": // Toggle selection
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

	// Pass other messages to datasets list
	var cmd tea.Cmd
	m.datasetsList, cmd = m.datasetsList.Update(msg)
	return m, cmd
}

// Update for dataset configuration state
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
			threshold, _ := strconv.Atoi(m.thresholdInput.Value())
			idx := m.datasetsList.Index()
			items := m.datasetsList.Items()
			item := items[idx].(DatasetItem)
			item.Threshold = threshold
			item.Unit = m.unitInput.Value()
			items[idx] = item
			m.datasetsList.SetItems(items)
			m.updateDatasetConfigs()
			m.state = StateSelectingDatasets
			return m, nil

		case "esc", "":
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
