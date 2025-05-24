package run

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

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
	if slices.Contains(m.installedIntegrations, integrationName) {
		return nil
	}

	integrationVersion, err := m.GetLatestPkgVersion(integrationName)
	if err != nil {
		return err
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
