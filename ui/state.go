package ui

// AppState holds the shared state between tabs
type AppState struct {
	selectedIntegrations map[string]bool
	datasetConfigs       map[string][]DatasetConfig
}

// Create a new AppState
func NewAppState() *AppState {
	return &AppState{
		selectedIntegrations: make(map[string]bool),
		datasetConfigs:       make(map[string][]DatasetConfig),
	}
}

type DatasetConfig struct {
	Name      string
	Selected  bool
	Threshold int
	Unit      string // "eps" or "bytes"
}
