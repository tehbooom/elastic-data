package run

import (
	"context"
	"encoding/json"
	"math/rand"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/tehbooom/elastic-data/internal/elasticsearch"
	"github.com/tehbooom/elastic-data/internal/generator"
	programContext "github.com/tehbooom/elastic-data/ui/context"
)

type DataGenerator struct {
	integrationName  string
	config           programContext.DatasetConfig
	ctx              context.Context
	cancel           context.CancelFunc
	stats            *IntegrationStats
	wg               *sync.WaitGroup
	mu               sync.RWMutex
	client           *elasticsearch.Config
	templates        []generator.LogTemplate
	dataPools        *generator.DataPools
	randGen          *rand.Rand
	bytesSent        int
	averageEventSize int
	fieldPatterns    map[string]*regexp.Regexp
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

		integrationDatasets := m.programContext.DatasetConfigs[integrationName]

		if dataset, ok := integrationDatasets[datasetName]; ok {
			templates, err := generator.LoadTemplatesForDataset(integrationName, datasetName)
			if err != nil {
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
				stats:            &stats,
				wg:               &m.wg,
				templates:        templates,
				client:           m.programContext.ESClient,
				averageEventSize: calculateAverageEventSize,
				integrationName:  integrationName,
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
	if dg.config.Unit == "bytes" {
		dg.sendBytes()
	}
	return nil
}

func (dg *DataGenerator) sendEPS() error {
	batchSize := dg.calculateOptimalBatchSize()
	var events []map[string]interface{}
	var batchBytes int
	for i := 0; i < batchSize; i++ {

		template := dg.templates[rand.Intn(len(dg.templates))]
		template.UpdateValues()

		message, err := template.ExecuteTemplate()
		if err != nil {
			return err
		}

		event := map[string]interface{}{
			"message": message,
		}

		events = append(events, event)

		eventBytes := dg.calculateEventSize(event)
		if dg.config.Unit == "bytes" && (dg.bytesSent+batchBytes+eventBytes) > dg.config.Threshold {
			break
		}
	}

	err := dg.sendBulkRequest(events)
	if err != nil {
		return err
	}

	dg.bytesSent += batchBytes
	return nil
}

func (dg *DataGenerator) sendBytes() error {
	batchSize := dg.calculateOptimalBatchSize()
	var events []map[string]interface{}
	var batchBytes int

	for i := 0; i < batchSize; i++ {
		if dg.config.Unit == "bytes" && dg.bytesSent >= dg.config.Threshold {
			break
		}
		template := dg.templates[rand.Intn(len(dg.templates))]
		template.UpdateValues()

		message, err := template.ExecuteTemplate()
		if err != nil {
			return err
		}

		event := map[string]interface{}{
			"message": message,
		}

		eventBytes := dg.calculateEventSize(event)
		if dg.config.Unit == "bytes" && (dg.bytesSent+batchBytes+eventBytes) > dg.config.Threshold {
			break
		}

		events = append(events, event)
		batchBytes += eventBytes
	}

	if len(events) == 0 {
		return nil
	}

	err := dg.sendBulkRequest(events)
	if err != nil {
		return err
	}

	dg.bytesSent += batchBytes
	return nil
}

func (dg *DataGenerator) calculateOptimalBatchSize() int {
	if dg.config.Unit == "eps" {
		target := dg.config.Threshold
		if target <= 10 {
			return 1
		} else if target <= 100 {
			return 10
		} else if target <= 1000 {
			return 100
		} else {
			return 500
		}
	} else {
		remainingBytes := dg.config.Threshold - dg.bytesSent

		if remainingBytes <= 0 {
			return 0
		}

		estimatedEvents := remainingBytes / dg.averageEventSize

		if estimatedEvents <= 0 {
			return 1
		} else if estimatedEvents <= 10 {
			return int(estimatedEvents)
		} else if estimatedEvents <= 100 {
			return 100
		} else {
			return 500
		}
	}
}

func (dg *DataGenerator) calculateEventSize(event map[string]interface{}) int {
	jsonBytes, _ := json.Marshal(event)
	return int(len(jsonBytes))
}

func (dg *DataGenerator) sendBulkRequest(events []map[string]interface{}) error {
	index := "logs-" + dg.integrationName + "." + dg.config.Name + "-default"
	dg.client.BulkRequest(index, events)
	return nil
}
