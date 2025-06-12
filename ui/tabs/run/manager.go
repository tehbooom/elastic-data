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
	m.mu.Lock()
	defer m.mu.Unlock()

	newIntegrations := make(map[string]*IntegrationStats)
	validIntegrations := make(map[string]bool)

	m.integrations = make(map[string]*IntegrationStats)

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
			validIntegrations[fullName] = true

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
			newIntegrations[fullName] = stats
		}
	}

	for fullName, generator := range m.generators {
		if !validIntegrations[fullName] {
			generator.stop()
			generator.resetStats()
			delete(m.generators, fullName)
		}
	}

	m.integrations = newIntegrations
}

func (m *TabModel) InstallPackage(integrationName string) error {
	if slices.Contains(m.installedIntegrations, integrationName) {
		log.Debug(fmt.Sprintf("Package %s installed", integrationName))
		return nil
	}

	log.Debug(fmt.Sprintf("Installing Package %s", integrationName))
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

		err := m.InstallPackage(integrationName)
		if err != nil {
			log.Debug(err)
			return err
		}

		integrationDatasets := m.programContext.DatasetConfigs[integrationName]

		if dataset, ok := integrationDatasets[datasetName]; ok {
			templates, err := generator.LoadTemplatesForDataset(m.programContext.ConfigPath, integrationName, datasetName, m.programContext.Config)
			if err != nil {
				log.Debug(err)
				return err
			}

			if len(templates) == 0 {
				return fmt.Errorf("loaded 0 templates for %s", fullName)
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
	for _, generator := range m.generators {
		generator.stop()
		generator.resetStats()
	}
	m.wg.Wait()
}

func (dg *DataGenerator) resetStats() {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	if dg.stats != nil {
		dg.stats.mu.Lock()
		dg.stats.Current = 0
		dg.stats.Peak = 0
		dg.stats.SentBytes = 0
		dg.stats.SentEvents = 0
		dg.stats.SentBytesUnit = ""
		dg.stats.Trend = "stable"
		dg.stats.mu.Unlock()
	}

	dg.bytesSent = 0
	dg.eventsSent = 0
	dg.averageEventSize = 0
}
