package context

import (
	"path/filepath"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/tehbooom/elastic-data/internal/config"
	"github.com/tehbooom/elastic-data/internal/elasticsearch"
	"github.com/tehbooom/elastic-data/internal/kibana"
)

type ProgramContext struct {
	Config               *config.Config
	ConfigPath           string
	SelectedIntegrations map[string]bool
	DatasetConfigs       map[string]map[string]DatasetConfig
	Dirty                bool
	ESClient             *elasticsearch.Config
	KBClient             *kibana.Config
	Error                error
	Running              bool
	mu                   sync.RWMutex
}

type DatasetConfig struct {
	Name                  string
	Selected              bool
	Threshold             int
	Unit                  string
	PreserveEventOriginal bool
	Events                []string
}

func NewProgramContext() *ProgramContext {
	return &ProgramContext{
		SelectedIntegrations: make(map[string]bool),
		DatasetConfigs:       make(map[string]map[string]DatasetConfig),
		Dirty:                false,
		Running:              false,
	}
}

func (a *ProgramContext) GetEnabledIntegrations() []string {
	var integrations []string
	for integration, enabled := range a.SelectedIntegrations {
		if enabled {
			integrations = append(integrations, integration)
		}
	}
	return integrations
}

func (a *ProgramContext) SetIntegrationSelected(integration string, selected bool) {
	a.SelectedIntegrations[integration] = selected
	a.Dirty = true
}

// SaveIntegrations persists the current integration and dataset selections to the config file.
// Only saves integrations that are either:
// - Enabled (regardless of dataset configuration)
// - Disabled but have existing dataset configurations (to preserve user's work)
// Disabled integrations with no datasets are omitted as they represent default state.
// For enabled integrations, only saves datasets with non-default configurations
func (a *ProgramContext) SaveIntegrations() {
	if a.Config == nil {
		log.Debug("Cannot save: no config loaded")
		return
	}
	updatedIntegrations := make(map[string]config.Integration)

	for integration, enabled := range a.SelectedIntegrations {
		if enabled {
			datasetConfigs, exists := a.DatasetConfigs[integration]
			datasetsToSave := make(map[string]config.Dataset)

			if exists {
				for datasetName, datasetConfig := range datasetConfigs {
					hasNonDefaultValues := datasetConfig.Threshold != 0 ||
						datasetConfig.Unit != "eps" ||
						datasetConfig.PreserveEventOriginal

					wasPreviouslyEnabled := false
					if existingIntegration, exists := a.Config.Integrations[integration]; exists {
						if existingDataset, exists := existingIntegration.Datasets[datasetName]; exists {
							wasPreviouslyEnabled = existingDataset.Enabled
						}
					}

					if datasetConfig.Selected || hasNonDefaultValues || wasPreviouslyEnabled {
						datasetsToSave[datasetName] = config.Dataset{
							Enabled:               datasetConfig.Selected,
							Threshold:             datasetConfig.Threshold,
							Unit:                  datasetConfig.Unit,
							Events:                datasetConfig.Events,
							PreserveEventOriginal: datasetConfig.PreserveEventOriginal,
						}
					}
				}
			}

			// Only save the integration if it has datasets worth saving
			if len(datasetsToSave) > 0 {
				updatedIntegrations[integration] = config.Integration{
					Enabled:  true,
					Datasets: datasetsToSave,
				}
			}
		} else {
			// For disabled integrations, only save if they have existing datasets
			existingIntegration, exists := a.Config.Integrations[integration]
			if exists && len(existingIntegration.Datasets) > 0 {
				updatedIntegrations[integration] = config.Integration{
					Enabled:  false,
					Datasets: existingIntegration.Datasets,
				}
			}
		}
	}

	a.Config.Integrations = updatedIntegrations
	if err := config.SaveConfig(a.Config, a.ConfigPath); err != nil {
		log.Printf("Error saving config: %v", err)
	}
	a.Dirty = false
}

func (a *ProgramContext) LoadFromConfig(cfg *config.Config, path string) {
	a.Config = cfg
	a.ConfigPath = filepath.Join(path, "config.yaml")

	if cfg.Integrations != nil {
		for integration, integrationData := range cfg.Integrations {
			a.SelectedIntegrations[integration] = integrationData.Enabled

			datasetMap, exists := a.DatasetConfigs[integration]
			if !exists {
				datasetMap = make(map[string]DatasetConfig)
				a.DatasetConfigs[integration] = datasetMap
			}

			for datasetName, configDataset := range integrationData.Datasets {
				datasetMap[datasetName] = DatasetConfig{
					Name:      datasetName,
					Selected:  configDataset.Enabled,
					Threshold: configDataset.Threshold,
					Unit:      configDataset.Unit,
					Events:    configDataset.Events,
				}
			}
		}
	}

	a.Dirty = false
}

func (pc *ProgramContext) IsRunning() bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.Running
}

func (pc *ProgramContext) SetRunning(running bool) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.Running = running
}
