package run

import (
	"context"
	"math/rand"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/tehbooom/elastic-data/internal/elasticsearch"
	"github.com/tehbooom/elastic-data/internal/generator"
	"github.com/tehbooom/elastic-data/ui/state"
)

type DataGenerator struct {
	config        state.DatasetConfig
	ctx           context.Context
	cancel        context.CancelFunc
	stats         *IntegrationStats
	wg            *sync.WaitGroup
	mu            sync.RWMutex
	data          string
	client        *elasticsearch.Config
	templates     []generator.LogTemplate
	dataPools     *generator.DataPools
	randGen       *rand.Rand
	fieldPatterns map[string]*regexp.Regexp
}

func (m *TabModel) StartGeneration() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stopAllGenerators()

	for fullName, stats := range m.integrations {
		fullNameSplit := strings.Split(fullName, ":")
		integrationName := fullNameSplit[0]
		datasetName := fullNameSplit[1]

		m.InstallPackage(integrationName)

		integrationDatasets := m.appState.DatasetConfigs[integrationName]

		if dataset, ok := integrationDatasets[datasetName]; ok {
			templates, err := generator.LoadTemplatesForDataset(integrationName, datasetName)
			if err != nil {
				return err
			}
			_ = templates

			ctx, cancel := context.WithCancel(m.mainCtx)

			generator := &DataGenerator{
				config: dataset,
				ctx:    ctx,
				cancel: cancel,
				stats:  &stats,
				wg:     &m.wg,
				client: m.appState.ESClient,
			}

			m.generators[fullName] = generator
			m.wg.Add(1)
			go generator.Start()
		}
	}
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

func (m *TabModel) stopGeneration() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopAllGenerators()
}

func (dg *DataGenerator) stop() {
	dg.cancel()
}

func (m *TabModel) stopAllGenerators() {
	for k, generator := range m.generators {
		generator.stop()
		delete(m.generators, k)
	}
	m.wg.Wait()
}

func (dg *DataGenerator) generateAndSendData() error {
	dg.mu.Lock()
	defer dg.mu.Unlock()
	return nil
}
