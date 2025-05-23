package run

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/tehbooom/elastic-data/internal/elasticsearch"
	"github.com/tehbooom/elastic-data/ui/state"
	"gopkg.in/yaml.v3"
)

type DataGenerator struct {
	config state.DatasetConfig
	ctx    context.Context
	cancel context.CancelFunc
	stats  *IntegrationStats
	wg     *sync.WaitGroup
	mu     sync.RWMutex
	data   string
	client *elasticsearch.Config
}

func (m *TabModel) StartGeneration() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for fullName, stats := range m.integrations {
		fullNameSplit := strings.Split(fullName, ":")
		integrationName := fullNameSplit[0]
		datasetName := fullNameSplit[1]

		m.InstallPackage(integrationName)
		integration := m.appState.DatasetConfigs[integrationName]
		dataset := integration[datasetName]

		ctx, cancel := context.WithCancel(m.mainCtx)

		generator := &DataGenerator{
			config: dataset,
			ctx:    ctx,
			cancel: cancel,
			stats:  &stats,
			wg:     &m.wg,
		}

		m.generators[fullName] = generator
		m.wg.Add(1)
		go generator.Start()
	}
	// for each m.intergations
	// get enabled datasets and their configs
	// for each dataset create psuedo data
	// start sending the psuedo data based on the config (threshold and unit)
	// if m.Running{
	// stop the running integrations from before
	// }
	return nil
}

func (dg *DataGenerator) Start() {
	defer dg.wg.Done()

	ticker := time.NewTicker(time.Duration(10 * time.Second))

	for {
		select {
		case <-dg.ctx.Done():
			log.Printf("Stopping data generation for %s", dg.config.Name)
			return
		case <-ticker.C:
			if err := dg.generateAndSendData(); err != nil {
				log.Printf("Error generating data for %s: %v", dg.config.Name, err)
			}
		}
	}
}

func (dg *DataGenerator) generateAndSendData() error {
	dg.mu.Lock()
	defer dg.mu.Unlock()
	data := dg.generateFakeData()
	return dg.index(data)
}

func (dg *DataGenerator) generateFakeData() map[string]interface{} {
	return make(map[string]interface{})
}

func (dg *DataGenerator) index(data any) error {
	dg.client.Client.Bulk().Index("")

	return nil
}

func getTrendIndicator(trend string) string {
	switch trend {
	case "up":
		return "↑"
	case "down":
		return "↓"
	default:
		return "─"
	}
}

// UpdateStats updates the statistics for the given integration
func (m *TabModel) UpdateStats(name string, value float64, unit string) {
	if m.integrations == nil {
		m.integrations = make(map[string]IntegrationStats)
	}

	integrationName := strings.Split(name, ":")

	stat, exists := m.integrations[name]
	if !exists {
		stat = IntegrationStats{
			Current:   value,
			Peak:      value,
			LastValue: value,
			Unit:      unit,
			Trend:     "stable",
		}
	} else {
		trend := "stable"
		if value > stat.LastValue {
			trend = "up"
		} else if value < stat.LastValue {
			trend = "down"
		}

		stat.Current = value
		stat.LastValue = value
		stat.Trend = trend

		if value > stat.Peak {
			stat.Peak = value
		}
	}

	if len(integrationName) > 1 {
		if configs, ok := m.appState.DatasetConfigs[integrationName[0]]; ok {
			if datasetConfig, ok := configs[integrationName[1]]; ok {
				stat.Unit = datasetConfig.Unit
			}
		}
	}

	if stat.Unit == "" {
		stat.Unit = unit
	}

	m.integrations[name] = stat
}

// RefreshIntegrations initializes or refreshes the integrations data from AppState
func (m *TabModel) RefreshIntegrations() {
	if m.integrations == nil {
		m.integrations = make(map[string]IntegrationStats)
	}

	m.integrations = make(map[string]IntegrationStats)

	for integrationName, isSelected := range m.appState.SelectedIntegrations {
		if !isSelected {
			continue
		}

		datasetConfigs, exists := m.appState.DatasetConfigs[integrationName]
		if !exists {
			continue
		}

		for datasetName, config := range datasetConfigs {
			if !config.Selected {
				continue
			}

			fullName := fmt.Sprintf("%s:%s", integrationName, datasetName)

			stats, exists := m.integrations[fullName]
			if !exists {
				stats = IntegrationStats{
					Current:   0,
					Peak:      0,
					LastValue: 0,
					Unit:      config.Unit,
					Trend:     "stable",
				}
			}

			stats.Unit = config.Unit

			m.integrations[fullName] = stats
		}
	}
}

func (m *TabModel) InstallPackage(integrationName string) error {
	integrationVersion, err := m.GetLatestPkgVersion(integrationName)
	if err != nil {
		return err
	}

	if slices.Contains(m.installedIntegrations, integrationName) {
		return nil
	}

	err = m.appState.KBClient.InstallPackage(integrationName, integrationVersion)
	if err != nil {
		return err
	}

	m.installedIntegrations = append(m.installedIntegrations, integrationName)

	return nil
}

func (m *TabModel) GetLatestPkgVersion(pkgName string) (string, error) {
	type Change struct {
		Description string `yaml:"description" json:"description"`
		Type        string `yaml:"type" json:"type"`
		Link        string `yaml:"link" json:"link"`
	}

	type Version struct {
		Version string   `yaml:"version" json:"version"`
		Changes []Change `yaml:"changes" json:"changes"`
	}

	type Changelog []Version

	var version string
	integrationPath := filepath.Join(m.appState.ConfigPath, "integrations", "packages", pkgName, "changelog.yml")

	file, err := os.ReadFile(integrationPath)
	if err != nil {
		return version, err
	}

	var changelog Changelog

	err = yaml.Unmarshal(file, &changelog)
	if err != nil {
		return version, err
	}

	version = changelog[0].Version

	return version, nil
}
