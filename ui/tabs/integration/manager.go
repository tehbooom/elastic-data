package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

func (m *TabModel) loadDatasetsForIntegration(integration string) error {
	var datasetItems []list.Item

	datasetMap, exists := m.context.DatasetConfigs[integration]
	if !exists {
		datasetMap = make(map[string]context.DatasetConfig)
		m.context.DatasetConfigs[integration] = datasetMap
	}

	repoDir := filepath.Join(m.context.ConfigPath, "integrations")

	log.Debug(fmt.Sprintf("path is %s", repoDir))

	dataSets, err := integrations.GetDatasets(repoDir, integration)
	if err != nil {
		return err
	}

	for i, j := range dataSets {
		log.Debug(fmt.Sprintf("Datset number %d: %s", i, j))
	}

	m.datasetsList.SetSize(m.width, len(dataSets)+2)

	for _, ds := range dataSets {
		existingConfig, configExists := datasetMap[ds]
		if !configExists {
			datasetMap[ds] = context.DatasetConfig{
				Name:                  ds,
				Selected:              false,
				Threshold:             0,
				Unit:                  "eps",
				PreserveEventOriginal: false,
				Events:                []string{},
			}
			existingConfig = datasetMap[ds]
		}
		datasetItems = append(datasetItems, NewDatasetItem(
			existingConfig.Name,
			existingConfig.Selected,
			existingConfig.Threshold,
			existingConfig.Unit,
			existingConfig.PreserveEventOriginal,
			existingConfig.Events,
		))
	}

	m.datasetsList.SetItems(datasetItems)
	m.datasetsList.Title = fmt.Sprintf("%s Datasets", strings.ToUpper(integration))
	return nil
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
			Name:                  datasetItem.Name,
			Selected:              datasetItem.Selected,
			Threshold:             datasetItem.Threshold,
			Unit:                  datasetItem.Unit,
			PreserveEventOriginal: datasetItem.PreserveEventOriginal,
			Events:                datasetItem.Events,
		}

		datasetMap[datasetItem.Name] = config
	}

	m.saveController.MarkDirty()
}

func (m *TabModel) getReadMe() (string, error) {
	integrationReadMe := filepath.Join(m.context.ConfigPath, "integrations", "packages", m.currentIntegration, "_dev", "build", "docs", "README.md")
	content, err := os.ReadFile(integrationReadMe)
	if err != nil {
		return "", err
	}

	pattern := `\{\{[^}]*\}\}`
	re := regexp.MustCompile(pattern)
	cleanedContent := re.ReplaceAllString(string(content), "")

	return cleanedContent, nil
}

func (m *TabModel) filterItems(items []list.Item, query string) []list.Item {
	if query == "" {
		return items
	}

	var filtered []list.Item
	query = strings.ToLower(query)

	if m.onlySelected {
		for _, item := range items {
			if integrationItem, ok := item.(*IntegrationItem); ok {
				if strings.HasPrefix(strings.ToLower(integrationItem.Title()), query) ||
					strings.HasPrefix(strings.ToLower(integrationItem.Name), query) && integrationItem.Selected {
					filtered = append(filtered, item)
				}
			}
		}
	} else {
		for _, item := range items {
			if integrationItem, ok := item.(*IntegrationItem); ok {
				if strings.HasPrefix(strings.ToLower(integrationItem.Title()), query) ||
					strings.HasPrefix(strings.ToLower(integrationItem.Name), query) {
					filtered = append(filtered, item)
				}
			}
		}
	}

	return filtered
}

func (m *TabModel) viewSelected(items []list.Item) []list.Item {

	var filtered []list.Item

	for _, item := range items {
		if integrationItem, ok := item.(*IntegrationItem); ok {
			if integrationItem.Selected {
				filtered = append(filtered, item)
			}
		}
	}
	return filtered
}
