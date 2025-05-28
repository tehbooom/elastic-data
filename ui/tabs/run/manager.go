package run

import (
	"context"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/tehbooom/elastic-data/internal/generator"
	"slices"
	"strings"
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

// RefreshIntegrations initializes or refreshes the integrations data from AppState
func (m *TabModel) RefreshIntegrations() {
	if m.integrations == nil {
		m.integrations = make(map[string]*IntegrationStats)
	}

	for integrationName, isSelected := range m.programContext.SelectedIntegrations {
		if !isSelected {
			continue
		}

		datasetConfigs, exists := m.programContext.DatasetConfigs[integrationName]
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
				stats = &IntegrationStats{
					Current: 0,
					Peak:    0,
					Unit:    config.Unit,
					Trend:   "neutral",
				}
			}

			stats.Unit = config.Unit

			m.integrations[fullName] = stats
		}
	}
}

func (m *TabModel) InstallPackage(integrationName string) error {
	if slices.Contains(m.installedIntegrations, integrationName) {
		log.Debug("Package installed")
		return nil
	}

	log.Debug("Installing Package ", integrationName)
	err := m.programContext.KBClient.InstallPackage(integrationName)
	if err != nil {
		log.Debug(err)
		return err
	}

	m.installedIntegrations = append(m.installedIntegrations, integrationName)

	return nil
}

func (m *TabModel) StartGeneration() error {

	log.Debug("StartGeneration")
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stopAllGenerators()

	for fullName, stats := range m.integrations {
		fullNameSplit := strings.Split(fullName, ":")
		integrationName := fullNameSplit[0]
		datasetName := fullNameSplit[1]

		m.InstallPackage(integrationName)

		integrationDatasets := m.programContext.DatasetConfigs[integrationName]

		if dataset, ok := integrationDatasets[datasetName]; ok {
			templates, err := generator.LoadTemplatesForDataset(m.programContext.ConfigPath, integrationName, datasetName)
			if err != nil {
				log.Debug(err)
				return err
			}

			var templateSizesTotal int
			for _, template := range templates {
				templateSizesTotal += template.Size
			}

			calculateAverageEventSize := templateSizesTotal / len(templates)

			ctx, cancel := context.WithCancel(m.mainCtx)

			generator := &DataGenerator{
				config:           dataset,
				ctx:              ctx,
				cancel:           cancel,
				stats:            stats,
				wg:               &m.wg,
				templates:        templates,
				client:           m.programContext.ESClient,
				averageEventSize: calculateAverageEventSize,
				integrationName:  integrationName,
			}

			m.generators[fullName] = generator
			m.wg.Add(1)
			if generator.config.Unit == "eps" {
				go generator.startEPS()
			} else {
				go generator.startBytes()
			}
		}
	}
	return nil
}

func (m *TabModel) stopGeneration() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopAllGenerators()
}

func (m *TabModel) stopAllGenerators() {
	for k, generator := range m.generators {
		generator.stop()
		delete(m.generators, k)
	}
	m.wg.Wait()
}
