package run

import (
	"sync"
	"time"
)

type IntegrationStats struct {
	Sent          float64
	Current       float64
	Peak          float64
	Unit          string
	LastValue     float64
	Trend         string
	recentBatches []BatchInfo
	lastUpdate    time.Time
	mu            sync.RWMutex
}

type BatchInfo struct {
	Timestamp time.Time
	SizeMB    float64
	Events    int
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
