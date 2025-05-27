package run

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/log"
)

var (
	trendUpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	trendDownStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	trendStableStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("249"))
)

type StatsSnapshot struct {
	Current float64
	Peak    float64
	Trend   string
	Unit    string
}

func (m *TabModel) getStatsSnapshot() map[string]StatsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshot := make(map[string]StatsSnapshot)
	for integration, stat := range m.integrations {
		stat.mu.RLock()
		snapshot[integration] = StatsSnapshot{
			Current: stat.Current,
			Peak:    stat.Peak,
			Trend:   stat.Trend,
			Unit:    stat.Unit, // or get this from config
		}
		stat.mu.RUnlock()
	}
	return snapshot
}

func (m *TabModel) RunTable() *table.Table {
	headers := []string{"Integration", "Dataset", "Current", "Peak", "Trend"}

	statsSnapshot := m.getStatsSnapshot()

	var rows [][]string
	for integration, stat := range statsSnapshot {
		log.Debug(fmt.Sprintf("tabl peak is %f", stat.Peak))
		var currentValue string
		var peakValue string

		if stat.Unit == "eps" {
			unit := "events"
			currentValue = fmt.Sprintf("%d %s", int(stat.Current), unit)
			peakValue = fmt.Sprintf("%d %s", int(stat.Peak), unit)
		} else {
			unit := "MB/s"
			currentValue = fmt.Sprintf("%.2f %s", stat.Current, unit)
			peakValue = fmt.Sprintf("%.2f %s", stat.Peak, unit)
		}

		trendIndicator := getTrendIndicator(stat.Trend)

		integrationSplit := strings.Split(integration, ":")

		row := []string{integrationSplit[0], integrationSplit[1], currentValue, peakValue, trendIndicator}

		rows = append(rows, row)
	}

	headerStyle := lipgloss.NewStyle().Bold(false).Foreground(lipgloss.Color("240"))
	baseStyle := lipgloss.NewStyle()

	t := table.New().
		Width(m.width - 2).
		Height(m.height).
		Headers(headers...).
		Rows(rows...).
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("238"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return headerStyle
			}

			rowIndex := row - 1
			if rowIndex < 0 || rowIndex >= len(rows) {
				return baseStyle
			}

			even := row%2 == 0
			if even {
				return baseStyle.Foreground(lipgloss.Color("245"))
			}

			if col == 4 {
				switch m.integrations[rows[rowIndex][0]+":"+rows[rowIndex][1]].Trend {
				case "up":
					return trendUpStyle
				case "down":
					return trendDownStyle
				default:
					return trendStableStyle
				}
			}

			return baseStyle.Foreground(lipgloss.Color("252"))
		})

	return t
}

// View renders the tab
func (m *TabModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	if m.integrations == nil || len(m.integrations) == 0 {
		log.Debug("No integrations")
		return "No active integrations."
	}

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Padding(0, 1)

	if m.running {
		statusStyle = statusStyle.Foreground(lipgloss.Color("42")) // Green for running
	} else {
		statusStyle = statusStyle.Foreground(lipgloss.Color("208")) // Orange for waiting
	}

	statusDisplay := statusStyle.Render(m.status)

	m.table = m.RunTable()

	return lipgloss.JoinVertical(lipgloss.Left, statusDisplay, baseStyle.Render(m.table.String())+"\n")

}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))
