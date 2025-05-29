package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/log"
	"github.com/tehbooom/elastic-data/internal/integrations"
	"github.com/tehbooom/elastic-data/ui/context"
)

func (m *TabModel) SetIntegrations(integrationsList []string) {
	m.integrationList.SetItems([]list.Item{})
	var items []list.Item

	for _, integration := range integrationsList {
		isSelected := false
		if m.context != nil && m.context.SelectedIntegrations != nil {
			if selected, exists := m.context.SelectedIntegrations[integration]; exists {
				isSelected = selected
			}
		}

		item := NewIntegrationItem(integration, isSelected)
		items = append(items, item)
	}

	m.integrationList.SetItems(items)
	log.Debug(fmt.Sprintf("Number of integrations is %d", len(m.integrationList.Items())))
}

func (m *TabModel) GetSelectedIntegrations() map[string]bool {
	result := make(map[string]bool)
	for _, item := range m.integrationList.Items() {
		if i, ok := item.(*IntegrationItem); ok {
			result[i.Name] = i.Selected
		}
	}
	return result
}

func (m *TabModel) loadDatasetsForIntegration(integration string) {
	var datasetItems []list.Item

	datasetMap, exists := m.context.DatasetConfigs[integration]
	if !exists {
		datasetMap = make(map[string]context.DatasetConfig)
		m.context.DatasetConfigs[integration] = datasetMap
	}

	configDir, _ := getConfigDir()
	repoDir := filepath.Join(configDir, "integrations")
	dataSets, err := integrations.GetDatasets(repoDir, integration)
	if err != nil {
		log.Fatal(err)
	}

	for _, ds := range dataSets {
		existingConfig, configExists := datasetMap[ds]
		if !configExists {
			datasetMap[ds] = context.DatasetConfig{
				Name:      ds,
				Selected:  false,
				Threshold: 0,
				Unit:      "eps",
			}
		} else {
			datasetMap[ds] = existingConfig
		}
	}

	for _, config := range datasetMap {
		datasetItems = append(datasetItems, NewDatasetItem(
			config.Name,
			config.Selected,
			config.Threshold,
			config.Unit,
		))
	}

	m.datasetsList.SetItems(datasetItems)
	m.datasetsList.Title = fmt.Sprintf("%s Datasets", strings.ToUpper(integration))
}

func (m *TabModel) updateDatasetConfigs() {
	if m.context == nil || m.currentIntegration == "" {
		fmt.Printf("ERROR: Cannot update app state - context: %v, currentIntegration: %s\n",
			m.context != nil, m.currentIntegration)
		return
	}

	datasetMap, exists := m.context.DatasetConfigs[m.currentIntegration]
	if !exists {
		datasetMap = make(map[string]context.DatasetConfig)
		m.context.DatasetConfigs[m.currentIntegration] = datasetMap
	}

	for _, item := range m.datasetsList.Items() {
		datasetItem, ok := item.(DatasetItem)
		if !ok {
			continue
		}

		config := context.DatasetConfig{
			Name:      datasetItem.Name,
			Selected:  datasetItem.Selected,
			Threshold: datasetItem.Threshold,
			Unit:      datasetItem.Unit,
		}

		datasetMap[datasetItem.Name] = config
	}

	m.saveController.MarkDirty()
}

func getConfigDir() (string, error) {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get user home directory: %w", err)
		}
		configHome = filepath.Join(home, ".config")
	}

	configDir := filepath.Join(configHome, "elastic-data")

	return configDir, nil
}
