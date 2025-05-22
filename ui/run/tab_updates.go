package run

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg struct{}

func (m *TabModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:

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
				// Initiate connection to cluster
				err := m.appState.ESClient.TestConnection()
				if err != nil {
					return m, tea.Batch(
						func() tea.Msg {
							return TabError{Message: "Failed to connect to cluster", Err: err}
						},
					)
				}

				err = m.appState.KBClient.TestConnection()
				if err != nil {
					return m, tea.Batch(
						func() tea.Msg {
							return TabError{Message: "Failed to connect to cluster", Err: err}
						},
					)
				}
			}

			// for each package do
			// install package
			// get psuedo data
			// send request for data

			return m, nil
		}
	case error:
		m.error = msg
		return m, tea.Quit
	case TabError:
		m.error = msg
		return m, tea.Quit
	}

	return m, nil
}
