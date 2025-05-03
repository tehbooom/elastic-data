package ui

import "github.com/tehbooom/elastic-data/internal/config"

// AppState holds the shared state between tabs
type AppState struct {
	selectedIntegrations map[string]bool
	datasetConfigs       map[string][]map[string]DatasetConfig
	configDir            string
	config               *config.Config
}

//	type Integration struct {
//		Name     string
//		Enabled  bool
//		Datasets []Dataset
//	}
//
//	type Dataset struct {
//		Name      string
//		Enabled   bool
//		Threshold struct {
//			EPS   *int64 `yaml:"eps,omitempty"`
//			Bytes *int64 `yaml:"bytes,omitempty"`
//		}
//	}
//
// Create a new AppState
func NewAppState() *AppState {
	return &AppState{
		selectedIntegrations: make(map[string]bool),
		datasetConfigs:       make(map[string][]map[string]DatasetConfig),
	}
}

type DatasetConfig struct {
	Name      string
	Selected  bool
	Threshold int
	// EPS or bytes
	Unit string
}

func (a *AppState) SaveIntegrations() {

}
