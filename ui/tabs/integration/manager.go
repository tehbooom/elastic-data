package integration

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/tehbooom/elastic-data/internal/integrations"
	"github.com/tehbooom/elastic-data/ui/context"
)

// SetIntegrations populates the integration list
func (m *TabModel) SetIntegrations(integrations []string) {
	m.integrationList.SetItems([]list.Item{})
	var items []list.Item

	for _, integration := range integrations {
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
}

// GetSelectedIntegrations returns a map of selected integrations
func (m *TabModel) GetSelectedIntegrations() map[string]bool {
	result := make(map[string]bool)
	for _, item := range m.integrationList.Items() {
		if i, ok := item.(*IntegrationItem); ok {
			result[i.Name] = i.Selected
		}
	}
	return result
}

// loadDatasetsForIntegration loads the datasets for the selected integration
func (m *TabModel) loadDatasetsForIntegration(integration string) {
	var datasetItems []list.Item

	// Check if we already have dataset configs for this integration
	datasetMap, exists := m.context.DatasetConfigs[integration]
	if !exists {
		// If not, create a new map for this integration
		datasetMap = make(map[string]context.DatasetConfig)
		m.context.DatasetConfigs[integration] = datasetMap
	}

	// Always get the datasets from the repo to ensure we have the complete list
	configDir, _ := getConfigDir()
	repoDir := filepath.Join(configDir, "integrations")
	dataSets, err := integrations.GetDatasets(repoDir, integration)
	if err != nil {
		log.Fatal(err)
	}

	// Process all datasets from the repository
	for _, ds := range dataSets {
		// Check if we already have config for this dataset
		existingConfig, configExists := datasetMap[ds]
		if !configExists {
			// If not, create a default config
			datasetMap[ds] = context.DatasetConfig{
				Name:      ds,
				Selected:  false,
				Threshold: 0,
				Unit:      "eps",
			}
		} else {
			// Keep the existing config
			datasetMap[ds] = existingConfig
		}
	}

	// Create list items for all datasets
	for _, config := range datasetMap {
		datasetItems = append(datasetItems, NewDatasetItem(
			config.Name,
			config.Selected,
			config.Threshold,
			config.Unit,
		))
	}

	// Update the UI list
	m.datasetsList.SetItems(datasetItems)
	m.datasetsList.Title = fmt.Sprintf("%s Datasets", strings.ToUpper(integration))
}

// updateDatasetConfigs updates the app state with the current dataset configurations
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

// getConfigDir returns the path to the configuration directory
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
