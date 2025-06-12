package run

import (
	"math"
	"sync"
	"time"
)

type IntegrationStats struct {
	// SentBytes amount of bytes sent
	SentBytes float64
	// SentBytesUnit unit of storage values are b, KB, MB,
	// GB, TB, PB, EB, ZB, YB
	SentBytesUnit string
	// SentEvents number of events sent for this integration
	SentEvents int
	// Current the latency in milliseconds for each bulk request to Elasticsearch
	Current float64
	// Peak the largest latency spike in milliseconds for bulk request to Elasticsearch
	Peak float64
	// Unit eps or bytes
	Unit string
	// Trend up down or neutral for the msot recent latency duration compared to the median
	Trend string
	// recentBatches a queue of recent bulk requests
	recentBatches []BatchInfo
	// lastUpdate time the stats were last updated
	lastUpdate time.Time
	mu         sync.RWMutex
}

type BatchInfo struct {
	Events   int
	Duration float64
}

func (stats *IntegrationStats) EnqueueRecentBatches(batch BatchInfo) {
	stats.recentBatches = append(stats.recentBatches, batch)

	if len(stats.recentBatches) > 11 {
		stats.recentBatches = stats.recentBatches[len(stats.recentBatches)-11:]
	}
}

func (stats *IntegrationStats) DequeueRecentBatches() (BatchInfo, bool) {
	if len(stats.recentBatches) == 0 {
		return BatchInfo{}, false
	}
	batch := stats.recentBatches[0]
	stats.recentBatches = stats.recentBatches[1:]
	return batch, true
}

func (stats *IntegrationStats) RecentBatchesSize() int {
	return len(stats.recentBatches)
}

func (stats *IntegrationStats) RecentBatchesIsEmpty() bool {
	return len(stats.recentBatches) == 0
}

// SetBytesUnit updates the SentBytes and SentBytesUnit for the stat
// calculating its unit of data storage.
func (stats *IntegrationStats) SetBytesUnit(b int) {
	bf := float64(b)

	for _, unit := range []string{"b", "KB", "MB", "GB", "TB", "PB", "EB", "ZB"} {
		if math.Abs(bf) < 1024.0 {
			stats.SentBytesUnit = unit
			stats.SentBytes = bf
			return
		}
		bf /= 1024.0
	}

	stats.SentBytesUnit = "YB"
	stats.SentBytes = bf
}

func (stats *IntegrationStats) CalculateLatency(duration time.Duration) {
	ms := float64(duration.Nanoseconds()) / 1e6
	stats.Current = ms

	if stats.Current > stats.Peak {
		stats.Peak = stats.Current
	}

	if len(stats.recentBatches) < 6 {
		stats.Trend = "neutral"
		return
	}

	batchCount := len(stats.recentBatches)

	// Recent average (last 3 batches including current)
	recentSum := stats.Current
	for i := batchCount - 2; i < batchCount && i >= 0; i++ {
		recentSum += stats.recentBatches[i].Duration
	}
	recentAvg := recentSum / 3

	// Older average (3 batches from earlier in the window)
	olderSum := 0.0
	startIdx := max(0, batchCount-6)
	for i := startIdx; i < startIdx+3 && i < batchCount; i++ {
		olderSum += stats.recentBatches[i].Duration
	}
	olderAvg := olderSum / 3

	threshold := olderAvg * 0.1
	if recentAvg > olderAvg+threshold {
		stats.Trend = "up"
	} else if recentAvg < olderAvg-threshold {
		stats.Trend = "down"
	} else {
		stats.Trend = "neutral"
	}
}
