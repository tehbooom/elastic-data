package run

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg struct{}

func (m *TabModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:

		for name, stat := range m.integrations {
			randomFactor := 0.9 + (0.2 * float64(time.Now().UnixNano()%100) / 100.0)
			newValue := stat.Current * randomFactor
			m.UpdateStats(name, newValue, stat.Unit)
		}
		return m, tea.Tick(time.Second, func(time.Time) tea.Msg {
			return tickMsg{}
		})

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			if m.running {
				m.running = false
				m.status = "Waiting to start"
			}
			return m, nil
		case "enter":
			if !m.running {
				m.running = true
				m.status = "Running"
			}
			// Initiate connection to cluster
			// for each package do
			// install package
			// get psuedo data
			// send request for data

			return m, nil
		}
	}

	return m, nil
}
