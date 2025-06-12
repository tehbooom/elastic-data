package run

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/log"
	"github.com/tehbooom/elastic-data/ui/style"
)

var (
	trendDownStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	trendUpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	trendStableStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("249"))
)

type StatsSnapshot struct {
	Current       float64
	Peak          float64
	Trend         string
	Unit          string
	SentBytes     float64
	SentBytesUnit string
	SentEvents    int
}

func (m *TabModel) getStatsSnapshot() map[string]StatsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshot := make(map[string]StatsSnapshot)
	for integration, generator := range m.generators {
		if generator.stats != nil {
			generator.stats.mu.RLock()
			snapshot[integration] = StatsSnapshot{
				Current:       generator.stats.Current,
				Peak:          generator.stats.Peak,
				Trend:         generator.stats.Trend,
				Unit:          generator.stats.Unit,
				SentBytes:     generator.stats.SentBytes,
				SentBytesUnit: generator.stats.SentBytesUnit,
				SentEvents:    generator.stats.SentEvents,
			}
			generator.stats.mu.RUnlock()
		}
	}

	for integration, stat := range m.integrations {
		if _, exists := snapshot[integration]; !exists {
			stat.mu.RLock()
			snapshot[integration] = StatsSnapshot{
				Current:       stat.Current,
				Peak:          stat.Peak,
				Trend:         stat.Trend,
				Unit:          stat.Unit,
				SentBytes:     stat.SentBytes,
				SentBytesUnit: stat.SentBytesUnit,
				SentEvents:    stat.SentEvents,
			}
			stat.mu.RUnlock()
		}
	}

	return snapshot
}

func (m *TabModel) RunTable() *table.Table {
	headers := []string{"Integration", "Dataset", "Sent", "Current", "Peak", "Trend"}
	statsSnapshot := m.getStatsSnapshot()

	var integrationNames []string
	for integration := range statsSnapshot {
		integrationNames = append(integrationNames, integration)
	}
	slices.Sort(integrationNames)

	var rows [][]string
	for _, integration := range integrationNames {
		stat := statsSnapshot[integration]
		currentValue := formatLatencyAdaptive(stat.Current)
		peakValue := formatLatencyAdaptive(stat.Peak)
		var sent string

		if stat.Unit == "eps" {
			sent = fmt.Sprintf("%d events", stat.SentEvents)
		} else {
			switch stat.SentBytesUnit {
			case "YB":
				sent = fmt.Sprintf("%.1f%s", stat.SentBytes, stat.SentBytesUnit)
			case "b":
				sent = fmt.Sprintf("%f bytes", stat.SentBytes)
			case "":
				sent = "0 bytes"
			default:
				sent = fmt.Sprintf("%3.1f%s", stat.SentBytes, stat.SentBytesUnit)
			}
		}

		var styledTrendIndicator string
		trendIndicator := getTrendIndicator(stat.Trend)
		switch stat.Trend {
		case "up":
			styledTrendIndicator = trendUpStyle.Render(trendIndicator)
		case "down":
			styledTrendIndicator = trendDownStyle.Render(trendIndicator)
		default:
			styledTrendIndicator = trendStableStyle.Render(trendIndicator)
		}

		integrationSplit := strings.Split(integration, ":")

		row := []string{integrationSplit[0], integrationSplit[1], sent, currentValue, peakValue, styledTrendIndicator}

		rows = append(rows, row)
	}

	t := table.New().
		Width(m.width - 2).
		Height(m.height).
		Headers(headers...).
		Rows(rows...).
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238")))

	return t
}

func (m *TabModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	if len(m.integrations) == 0 {
		log.Debug("No integrations")
		return "No active integrations."
	}

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Width(m.width-2).
		Align(lipgloss.Center).
		Bold(true).
		Padding(0, 1)

	if m.programContext.Running {
		statusStyle = statusStyle.Foreground(lipgloss.Color("42"))
	} else {
		statusStyle = statusStyle.Foreground(lipgloss.Color("208"))
	}

	statusDisplay := statusStyle.Render(m.status)

	m.table = m.RunTable()
	help := style.FormatHelp(
		"(enter)", "Start/Stop",
		"(q)", "Stop",
		"(tab)", "Switch tabs",
		"(ctrl+c)", "Quit",
	)

	return lipgloss.JoinVertical(lipgloss.Left, "\n"+statusDisplay, baseStyle.Render(m.table.String())+"\n"+help)

}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("240"))

func formatLatencyAdaptive(ms float64) string {
	switch {
	case ms >= 1000:
		return fmt.Sprintf("%.2f s", ms/1000)
	case ms >= 100:
		return fmt.Sprintf("%.0f ms", ms)
	case ms >= 10:
		return fmt.Sprintf("%.1f ms", ms)
	default:
		return fmt.Sprintf("%.2f ms", ms)
	}
}
