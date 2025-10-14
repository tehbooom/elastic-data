package run

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
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
	templates        []*generator.LogTemplate
	bytesSent        int
	eventsSent       int
	averageEventSize int
}

func (dg *DataGenerator) startBytes() {
	defer dg.wg.Done()

	ticker := time.NewTicker(time.Duration(10 * time.Second))
	defer ticker.Stop()
	for {
		select {
		case <-dg.ctx.Done():
			log.Debug("Stopping data generation for %s", dg.config.Name)
			return
		case <-ticker.C:
			if err := dg.sendBytes(); err != nil {
				log.Debug(err)
				log.Debug("Error generating data for %s: %v", dg.config.Name, err)
			}
			if dg.bytesSent >= dg.config.Threshold {
				log.Debug("Reached byte threshold for %s", dg.config.Name)
				return
			}
		}
	}
}

func (dg *DataGenerator) stop() {
	dg.cancel()
}

func (dg *DataGenerator) startEPS() {
	defer dg.wg.Done()
	targetEPS := dg.config.Threshold
	batchSize := dg.calculateOptimalBatchSize()

	batchInterval := time.Duration(batchSize) * time.Second / time.Duration(targetEPS)
	log.Debug(fmt.Sprintf("Batch interval set to %v for %s", batchInterval, dg.config.Name))

	ticker := time.NewTicker(batchInterval)
	defer ticker.Stop()
	log.Debug("Starting EPS generation for %s: %d EPS (batch size: %d, interval: %v)",
		dg.config.Name, targetEPS, batchSize, batchInterval)

	// Send first batch immediately instead of waiting for first tick
	if err := dg.sendEPS(); err != nil {
		log.Debug(err)
		log.Debug("Error sending initial EPS batch for %s: %v", dg.config.Name, err)
	}

	for {
		select {
		case <-dg.ctx.Done():
			log.Debug("Stopping EPS generation for %s", dg.config.Name)
			return
		case <-ticker.C:
			if err := dg.sendEPS(); err != nil {
				log.Debug(err)
				log.Debug("Error sending EPS batch for %s: %v", dg.config.Name, err)
			}
		}
	}
}

func (dg *DataGenerator) sendEPS() error {
	// Get batch configuration without holding lock
	batchSize := dg.calculateOptimalBatchSize()

	dg.mu.Lock()
	selectedTemplates := dg.selectTemplatesAdaptive(batchSize)
	preserveOriginal := dg.config.PreserveEventOriginal
	dg.mu.Unlock()

	// Generate events outside of lock
	events := make([]map[string]interface{}, 0, batchSize)
	var batchBytes int
	timestamp := time.Now().UTC().Format(time.RFC3339)
	timestampOverhead := 50 // Approximate overhead for @timestamp and other metadata

	for i := 0; i < batchSize; i++ {
		template := selectedTemplates[i%len(selectedTemplates)]
		template.UpdateValues()

		message, err := template.ExecuteTemplate()
		if err != nil {
			log.Debug(err)
			return err
		}

		messageLen := len(message)
		var event map[string]interface{}

		if template.IsJSON {
			decoder := json.NewDecoder(strings.NewReader(message))
			if err := decoder.Decode(&event); err != nil {
				log.Debug("Failed to parse JSON message:", err)
				return err
			}
			event["@timestamp"] = timestamp
		} else {
			event = map[string]interface{}{
				"message":    message,
				"@timestamp": timestamp,
			}
		}

		if preserveOriginal {
			event["tags"] = []string{"preserve_original_event"}
		}

		log.Debug(fmt.Sprintf("Event %d: %v", i, event))

		events = append(events, event)

		// Approximate size without re-marshaling
		eventBytes := messageLen + timestampOverhead
		batchBytes += eventBytes
	}

	if len(events) == 0 {
		return nil
	}

	// Send bulk request without lock
	duration, err := dg.sendBulkRequest(events)
	if err != nil {
		log.Debug(err)
		return err
	}

	// Only lock for updating stats
	dg.mu.Lock()
	dg.bytesSent += batchBytes
	dg.updateStats(len(events), duration)
	dg.mu.Unlock()

	return nil
}

func (dg *DataGenerator) sendBytes() error {
	// Get current state and batch configuration with minimal lock time
	dg.mu.Lock()
	batchSize := dg.calculateOptimalBatchSize()
	currentBytesSent := dg.bytesSent
	threshold := dg.config.Threshold
	selectedTemplates := dg.selectTemplatesAdaptive(batchSize)
	preserveOriginal := dg.config.PreserveEventOriginal
	dg.mu.Unlock()

	log.Debug(fmt.Sprintf("Batch size is %d for %s", batchSize, dg.config.Name))

	// Generate events outside of lock
	events := make([]map[string]interface{}, 0, batchSize)
	var batchBytes int
	timestamp := time.Now().UTC().Format(time.RFC3339)
	timestampOverhead := 50 // Approximate overhead for @timestamp and other metadata

	for i := 0; i < batchSize; i++ {
		if dg.config.Unit == "bytes" && currentBytesSent >= threshold {
			log.Debug(fmt.Sprintf("Threshold %d met for %s", batchSize, dg.config.Name))
			break
		}
		template := selectedTemplates[i%len(selectedTemplates)]
		template.UpdateValues()

		message, err := template.ExecuteTemplate()
		if err != nil {
			log.Debug(err)
			return err
		}

		messageLen := len(message)
		var event map[string]interface{}

		if template.IsJSON {
			decoder := json.NewDecoder(strings.NewReader(message))
			if err := decoder.Decode(&event); err != nil {
				log.Debug("Failed to parse JSON message:", err)
				return err
			}
			event["@timestamp"] = timestamp
		} else {
			event = map[string]interface{}{
				"message":    message,
				"@timestamp": timestamp,
			}
		}

		if preserveOriginal {
			event["tags"] = []string{"preserve_original_event"}
		}

		// Approximate size without re-marshaling
		eventBytes := messageLen + timestampOverhead

		if dg.config.Unit == "bytes" && (currentBytesSent+batchBytes+eventBytes) > threshold {
			break
		}

		events = append(events, event)
		batchBytes += eventBytes
	}

	if len(events) == 0 {
		return nil
	}

	// Send bulk request without lock
	duration, err := dg.sendBulkRequest(events)
	if err != nil {
		log.Debug(err)
		return err
	}

	// Only lock for updating stats
	dg.mu.Lock()
	dg.bytesSent += batchBytes
	dg.updateStats(len(events), duration)
	dg.mu.Unlock()

	return nil
}

func (dg *DataGenerator) selectTemplatesAdaptive(batchSize int) []*generator.LogTemplate {
	var defaultTemplates []*generator.LogTemplate
	var userTemplates []*generator.LogTemplate

	for _, template := range dg.templates {
		if template.UserProvided {
			log.Debug("Found user provided template")
			userTemplates = append(userTemplates, template)
		} else {
			defaultTemplates = append(defaultTemplates, template)
		}
	}

	if len(userTemplates) == 0 {
		return dg.selectRandomTemplates(dg.templates, batchSize)
	}

	var userCount int

	switch {
	case len(userTemplates) == 1:
		userCount = 1
	case len(userTemplates) <= 3:
		userCount = len(userTemplates)
	case len(userTemplates) < batchSize/3:
		userCount = len(userTemplates)
	default:
		userCount = batchSize / 3
	}

	log.Debug(fmt.Sprintf("User count is %d", userCount))

	userCount = min(userCount, batchSize)

	result := make([]*generator.LogTemplate, 0, batchSize)

	if userCount > 0 {
		userSelected := dg.selectRandomTemplates(userTemplates, userCount)
		result = append(result, userSelected...)
	}

	remainingSlots := batchSize - len(result)
	if remainingSlots > 0 && len(defaultTemplates) > 0 {
		defaultSelected := dg.selectRandomTemplates(defaultTemplates, remainingSlots)
		result = append(result, defaultSelected...)
	}

	rand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})

	return result
}

func (dg *DataGenerator) selectRandomTemplates(templates []*generator.LogTemplate, count int) []*generator.LogTemplate {
	if count >= len(templates) {
		result := make([]*generator.LogTemplate, len(templates))
		copy(result, templates)
		return result
	}

	indices := make([]int, len(templates))
	for i := range indices {
		indices[i] = i
	}
	rand.Shuffle(len(indices), func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})

	result := make([]*generator.LogTemplate, count)
	for i := 0; i < count; i++ {
		result[i] = templates[indices[i]]
	}

	return result
}

func (dg *DataGenerator) calculateOptimalBatchSize() int {
	if dg.config.Unit == "eps" {
		target := dg.config.Threshold
		// Larger batches for higher EPS to reduce overhead
		if target <= 10 {
			return 1
		} else if target <= 50 {
			return 10
		} else if target <= 200 {
			return 50
		} else if target <= 1000 {
			return 200
		} else if target <= 5000 {
			return 1000
		} else {
			return 2000
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
		} else if estimatedEvents <= 1000 {
			return 1000
		} else {
			return 2000
		}
	}
}

func (dg *DataGenerator) calculateEventSize(event map[string]interface{}) int {
	jsonBytes, _ := json.Marshal(event)
	return int(len(jsonBytes))
}

func (dg *DataGenerator) sendBulkRequest(events []map[string]interface{}) (time.Duration, error) {
	index := "logs-" + dg.integrationName + "." + dg.config.Name + "-default"
	duration, err := dg.client.BulkRequest(index, events)
	if err != nil {
		log.Debug(err)
		return duration, err
	}
	return duration, nil
}

func (dg *DataGenerator) updateStats(eventCount int, duration time.Duration) {
	if dg.stats == nil {
		return
	}
	dg.stats.mu.Lock()
	defer dg.stats.mu.Unlock()

	durationNano := float64(duration.Nanoseconds()) / 1e6

	dg.stats.SentEvents += eventCount

	if dg.stats.Unit == "bytes" {
		dg.stats.SetBytesUnit(dg.bytesSent)
	}

	dg.stats.CalculateLatency(duration)
	now := time.Now()
	dg.stats.EnqueueRecentBatches(BatchInfo{
		Events:   eventCount,
		Duration: durationNano,
	})
	dg.stats.lastUpdate = now
}
