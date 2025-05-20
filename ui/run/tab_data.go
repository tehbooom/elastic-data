package run

import (
	"fmt"
	"strings"

	"github.com/tehbooom/elastic-data/internal/elasticsearch"
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

func (m *TabModel) TestConnection() error {

	err := elasticsearch.TestConnection(m.appState.Config.Connection)
	if err != nil {
		return err
	}

	return nil
}
