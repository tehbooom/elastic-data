package ui

import (
	"fmt"
	"log"

	"github.com/tehbooom/elastic-data/internal/config"
)

// AppState holds the shared state between tabs
type AppState struct {
	ConfigPath           string
	Config               *config.Config
	DatasetConfigs       map[string]map[string]DatasetConfig
	Dirty                bool
	SelectedIntegrations map[string]bool
}

// Create a new AppState
func NewAppState() *AppState {
	return &AppState{
		SelectedIntegrations: make(map[string]bool),
		DatasetConfigs:       make(map[string]map[string]DatasetConfig),
		Dirty:                false,
	}
}

type DatasetConfig struct {
	Name      string
	Selected  bool
	Threshold int
	// EPS or bytes
	Unit string
}

// SaveIntegrations saves selectedIntegrations and datasetConfigs to the config field
func (a *AppState) SaveIntegrations() {
	if a.Config.Integrations == nil {
		a.Config.Integrations = make(map[string]config.Integration)
	}

	// Create a new map to rebuild from scratch
	updatedIntegrations := make(map[string]config.Integration)

	// Process all integrations that exist in our UI state
	for integration, enabled := range a.SelectedIntegrations {
		// For selected integrations, save their dataset configs
		if enabled {
			uiDatasets, exists := a.DatasetConfigs[integration]
			if !exists {
				continue
			}

			datasetsToSave := make(map[string]config.Dataset)
			hasConfiguredDatasets := false

			for datasetName, uiDataset := range uiDatasets {
				if uiDataset.Selected && (uiDataset.Threshold != 0 || uiDataset.Unit != "eps") {
					hasConfiguredDatasets = true
					datasetsToSave[datasetName] = config.Dataset{
						Enabled:   uiDataset.Selected,
						Threshold: uiDataset.Threshold,
						Unit:      uiDataset.Unit,
					}
				}
			}

			if hasConfiguredDatasets {
				updatedIntegrations[integration] = config.Integration{
					Enabled:  true,
					Datasets: datasetsToSave,
				}
			}
		}
		// If integration is not enabled, we don't add it to updatedIntegrations
	}

	// Replace the entire integrations map with our updated version
	a.Config.Integrations = updatedIntegrations

	// Save the config
	if err := config.SaveConfig(a.Config, a.ConfigPath); err != nil {
		log.Printf("Error saving config: %v", err)
	}
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

	// Reset dirty flag
	a.Dirty = false
}

// SaveToConfig saves the application state to the config
func (a *AppState) SaveToConfig() error {
	if a.Config == nil {
		return fmt.Errorf("no config loaded")
	}

	// Create updated integrations map
	updatedIntegrations := make(map[string]config.Integration)

	// Add selected integrations with configured datasets
	for integration, selected := range a.SelectedIntegrations {
		if selected {
			datasetsMap, exists := a.DatasetConfigs[integration]
			if !exists {
				continue
			}

			// Convert datasets to config format
			configDatasets := make(map[string]config.Dataset)
			hasConfiguredDatasets := false

			for datasetName, datasetConfig := range datasetsMap {
				if datasetConfig.Selected && (datasetConfig.Threshold != 0 || datasetConfig.Unit != "eps") {
					hasConfiguredDatasets = true
					configDatasets[datasetName] = config.Dataset{
						Enabled:   datasetConfig.Selected,
						Threshold: datasetConfig.Threshold,
						Unit:      datasetConfig.Unit,
					}
				}
			}

			if hasConfiguredDatasets {
				updatedIntegrations[integration] = config.Integration{
					Enabled:  true,
					Datasets: configDatasets,
				}
			}
		}
	}

	// Update config and save
	a.Config.Integrations = updatedIntegrations
	err := config.SaveConfig(a.Config, a.ConfigPath)
	if err == nil {
		a.Dirty = false
	}
	return err
}

// Methods to mark state as dirty
func (a *AppState) SetIntegrationSelected(integration string, selected bool) {
	a.SelectedIntegrations[integration] = selected
	a.Dirty = true
}

func (a *AppState) SetDatasetConfig(integration, dataset string, config DatasetConfig) {
	if _, exists := a.DatasetConfigs[integration]; !exists {
		a.DatasetConfigs[integration] = make(map[string]DatasetConfig)
	}
	a.DatasetConfigs[integration][dataset] = config
	a.Dirty = true
}
