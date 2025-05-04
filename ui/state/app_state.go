package state

import (
	"log"

	"github.com/tehbooom/elastic-data/internal/config"
)

// AppState holds the shared state between tabs
type AppState struct {
	Config               *config.Config
	ConfigPath           string
	SelectedIntegrations map[string]bool
	DatasetConfigs       map[string]map[string]DatasetConfig
	Dirty                bool
}

// DatasetConfig represents configuration for a dataset
type DatasetConfig struct {
	Name      string
	Selected  bool
	Threshold int
	Unit      string
}

// NewAppState creates a new application state
func NewAppState() *AppState {
	return &AppState{
		SelectedIntegrations: make(map[string]bool),
		DatasetConfigs:       make(map[string]map[string]DatasetConfig),
		Dirty:                false,
	}
}

// SetIntegrationSelected updates an integration's selection state
func (a *AppState) SetIntegrationSelected(integration string, selected bool) {
	a.SelectedIntegrations[integration] = selected
	a.Dirty = true
}

// SaveIntegrations saves the current state to the config
func (a *AppState) SaveIntegrations() {
	if a.Config == nil {
		log.Println("Cannot save: no config loaded")
		return
	}

	// Create updated integrations map
	updatedIntegrations := make(map[string]config.Integration)

	// Process all integrations that exist in our UI state
	for integration, enabled := range a.SelectedIntegrations {
		// For selected integrations, save their dataset configs
		if enabled {
			datasetConfigs, exists := a.DatasetConfigs[integration]

			// Initialize the datasets map
			datasetsToSave := make(map[string]config.Dataset)

			// If the integration has dataset configs, process them
			if exists {
				for datasetName, datasetConfig := range datasetConfigs {
					// Save datasets that are selected AND have non-default configuration
					if datasetConfig.Selected && (datasetConfig.Threshold != 0 || datasetConfig.Unit != "eps") {
						datasetsToSave[datasetName] = config.Dataset{
							Enabled:   datasetConfig.Selected,
							Threshold: datasetConfig.Threshold,
							Unit:      datasetConfig.Unit,
						}
					}
				}
			}

			// Always save selected integrations, even if no datasets are configured
			updatedIntegrations[integration] = config.Integration{
				Enabled:  true,
				Datasets: datasetsToSave,
			}
		}
	}

	// Replace the entire integrations map with our updated version
	a.Config.Integrations = updatedIntegrations

	// Save the config
	if err := config.SaveConfig(a.Config, a.ConfigPath); err != nil {
		log.Printf("Error saving config: %v", err)
	}

	a.Dirty = false
}

// LoadFromConfig loads the application state from the config
func (a *AppState) LoadFromConfig(cfg *config.Config, path string) {
	a.Config = cfg
	a.ConfigPath = path

	// Initialize from config
	if cfg.Integrations != nil {
		for integration, integrationData := range cfg.Integrations {
			// Set selection state
			a.SelectedIntegrations[integration] = integrationData.Enabled

			// Create dataset map if it doesn't exist
			datasetMap, exists := a.DatasetConfigs[integration]
			if !exists {
				datasetMap = make(map[string]DatasetConfig)
				a.DatasetConfigs[integration] = datasetMap
			}

			// Add datasets
			for datasetName, configDataset := range integrationData.Datasets {
				datasetMap[datasetName] = DatasetConfig{
					Name:      datasetName,
					Selected:  configDataset.Enabled,
					Threshold: configDataset.Threshold,
					Unit:      configDataset.Unit,
				}
			}
		}
	}

	a.Dirty = false
}
