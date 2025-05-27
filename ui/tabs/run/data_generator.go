package run

import (
	"context"
	"encoding/json"
	"fmt"
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

func (dg *DataGenerator) startBytes() {
	defer dg.wg.Done()

	ticker := time.NewTicker(time.Duration(10 * time.Second))
	defer ticker.Stop()
	for {
		select {
		case <-dg.ctx.Done():
			log.Printf("Stopping data generation for %s", dg.config.Name)
			return
		case <-ticker.C:
			if err := dg.sendBytes(); err != nil {
				log.Debug(err)
				log.Printf("Error generating data for %s: %v", dg.config.Name, err)
			}
			if dg.bytesSent >= dg.config.Threshold {
				log.Printf("Reached byte threshold for %s", dg.config.Name)
				return
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

func (dg *DataGenerator) startEPS() {
	defer dg.wg.Done()
	targetEPS := dg.config.Threshold
	batchSize := dg.calculateOptimalBatchSize()

	batchInterval := time.Duration(batchSize) * time.Second / time.Duration(targetEPS)
	log.Debug(fmt.Sprintf("Batch interval set to %d for %s", batchInterval, dg.config.Name))

	ticker := time.NewTicker(batchInterval)
	defer ticker.Stop()
	log.Printf("Starting EPS generation for %s: %d EPS (batch size: %d, interval: %v)",
		dg.config.Name, targetEPS, batchSize, batchInterval)

	for {
		select {
		case <-dg.ctx.Done():
			log.Printf("Stopping EPS generation for %s", dg.config.Name)
			return
		case <-ticker.C:
			if err := dg.sendEPS(); err != nil {
				log.Debug(err)
				log.Printf("Error sending EPS batch for %s: %v", dg.config.Name, err)
			}
		}
	}
}

func (dg *DataGenerator) sendEPS() error {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	batchSize := dg.calculateOptimalBatchSize()
	var events []map[string]interface{}
	var batchBytes int

	for i := 0; i < batchSize; i++ {
		template := dg.templates[rand.Intn(len(dg.templates))]
		template.UpdateValues()

		message, err := template.ExecuteTemplate()
		if err != nil {
			log.Debug(err)
			return err
		}

		event := map[string]interface{}{
			"message":    message,
			"@timestamp": time.Now().UTC().Format(time.RFC3339),
		}

		events = append(events, event)

		eventBytes := dg.calculateEventSize(event)
		batchBytes += eventBytes
	}

	if len(events) == 0 {
		return nil
	}

	err := dg.sendBulkRequest(events)
	if err != nil {
		log.Debug(err)
		return err
	}

	dg.bytesSent += batchBytes
	dg.updateStats(len(events), batchBytes)
	return nil
}

func (dg *DataGenerator) sendBytes() error {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	batchSize := dg.calculateOptimalBatchSize()
	log.Debug(fmt.Sprintf("Batch size is %d for %s", batchSize, dg.config.Name))

	var events []map[string]interface{}
	var batchBytes int

	for i := 0; i < batchSize; i++ {
		if dg.config.Unit == "bytes" && dg.bytesSent >= dg.config.Threshold {
			log.Debug(fmt.Sprintf("Threshold %d met for %s", batchSize, dg.config.Name))
			break
		}
		template := dg.templates[rand.Intn(len(dg.templates))]
		template.UpdateValues()

		message, err := template.ExecuteTemplate()
		if err != nil {
			log.Debug(err)
			return err
		}
		log.Debug(fmt.Sprintf("event %d for %s: %s", i, dg.config.Name, message))

		event := map[string]interface{}{
			"message":    message,
			"@timestamp": time.Now().UTC().Format(time.RFC3339),
		}

		eventBytes := dg.calculateEventSize(event)

		// Leave this commented incase we want to allow a hard stop at the threshold
		// if dg.config.Unit == "bytes" && (dg.bytesSent+batchBytes+eventBytes) > dg.config.Threshold {
		// 	break
		// }

		events = append(events, event)
		batchBytes += eventBytes
	}

	if len(events) == 0 {
		return nil
	}

	err := dg.sendBulkRequest(events)
	if err != nil {
		log.Debug(err)
		return err
	}

	dg.bytesSent += batchBytes
	dg.updateStats(len(events), batchBytes)
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

func (dg *DataGenerator) updateStats(eventCount, byteCount int) {
	if dg.stats == nil {
		return
	}

	dg.stats.mu.Lock()
	defer dg.stats.mu.Unlock()

	if dg.stats.Unit == "bytes" {
		sizeMB := float64(byteCount) / (1024 * 1024)
		now := time.Now()
		dg.stats.LastValue = dg.stats.Current
		dg.stats.Current += sizeMB
		dg.stats.recentBatches = append(dg.stats.recentBatches, BatchInfo{
			Timestamp: now,
			SizeMB:    sizeMB,
			Events:    eventCount,
		})
		cutoff := now.Add(-60 * time.Second)
		var validBatches []BatchInfo
		for _, batch := range dg.stats.recentBatches {
			if batch.Timestamp.After(cutoff) {
				validBatches = append(validBatches, batch)
			}
		}
		dg.stats.recentBatches = validBatches

		dg.stats.Peak = dg.stats.calculatePeakThroughput()

		dg.stats.Trend = dg.stats.calculateTrend()

		dg.stats.lastUpdate = now
		log.Debug(fmt.Sprintf("Current is %v for %s", dg.stats.Current, dg.config.Name))
		log.Debug(fmt.Sprintf("Peak is %v for %s", dg.stats.Peak, dg.config.Name))
		log.Debug(fmt.Sprintf("Trend is %v for %s", dg.stats.Trend, dg.config.Name))
	} else {
		now := time.Now()
		dg.stats.LastValue = dg.stats.Current
		dg.stats.recentBatches = append(dg.stats.recentBatches, BatchInfo{
			Timestamp: now,
			Events:    eventCount,
		})
		cutoff := now.Add(-60 * time.Second)
		var validBatches []BatchInfo
		for _, batch := range dg.stats.recentBatches {
			if batch.Timestamp.After(cutoff) {
				validBatches = append(validBatches, batch)
			}
		}
		dg.stats.recentBatches = validBatches
		dg.stats.Current = float64(eventCount)
		dg.stats.Peak = dg.stats.calculatePeakThroughput()
		dg.stats.Trend = dg.stats.calculateTrend()
	}
}

func (stats *IntegrationStats) calculatePeakThroughput() float64 {
	if len(stats.recentBatches) < 2 {
		return 0
	}

	var maxThroughput float64
	windowSize := 10 * time.Second

	for i := 0; i < len(stats.recentBatches); i++ {
		windowStart := stats.recentBatches[i].Timestamp
		windowEnd := windowStart.Add(windowSize)

		var windowBytes float64
		for j := i; j < len(stats.recentBatches); j++ {
			if stats.recentBatches[j].Timestamp.After(windowEnd) {
				break
			}
			windowBytes += stats.recentBatches[j].SizeMB
		}

		throughput := windowBytes / windowSize.Seconds()
		if throughput > maxThroughput {
			maxThroughput = throughput
		}
	}

	return maxThroughput
}

func (stats *IntegrationStats) calculateTrend() string {
	if len(stats.recentBatches) < 3 {
		return "neutral"
	}

	now := time.Now()
	recent := now.Add(-30 * time.Second)
	older := now.Add(-60 * time.Second)

	var recentMB, olderMB float64

	for _, batch := range stats.recentBatches {
		if batch.Timestamp.After(recent) {
			recentMB += batch.SizeMB
		} else if batch.Timestamp.After(older) {
			olderMB += batch.SizeMB
		}
	}

	if olderMB == 0 {
		if recentMB > 0 {
			return "up"
		}
		return "neutral"
	}

	change := (recentMB - olderMB) / olderMB

	if change > 0.1 {
		return "up"
	} else if change < -0.1 {
		return "down"
	}
	return "neutral"
}

func (stats *IntegrationStats) GetCurrentThroughput() float64 {
	stats.mu.RLock()
	defer stats.mu.RUnlock()

	now := time.Now()
	recent := now.Add(-10 * time.Second)

	var recentMB float64
	for _, batch := range stats.recentBatches {
		if batch.Timestamp.After(recent) {
			recentMB += batch.SizeMB
		}
	}

	return recentMB / 10.0
}
