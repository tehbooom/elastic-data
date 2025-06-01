package run

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

type tickMsg struct{}

type refreshMsg struct{}

func (m *TabModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		if !m.shouldTick {
			return m, nil
		}
		return m, tea.Tick(time.Second, func(time.Time) tea.Msg {
			return tickMsg{}
		})

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			if m.running {
				m.running = false
				m.status = "Stopping..."
				m.stopGeneration()
				m.shouldTick = false
				m.status = "Waiting to start"
			}
			return m, nil
		case "enter":
			if !m.running {
				m.running = true
				m.status = StartedMsg
				err := m.programContext.ESClient.TestConnection()
				if err != nil {
					log.Debug(err)
					return m, tea.Batch(
						func() tea.Msg {
							return TabError{Message: "Failed to connect to cluster", Err: err}
						},
					)
				}

				err = m.programContext.KBClient.TestConnection()
				if err != nil {
					log.Debug(err)
					return m, tea.Batch(
						func() tea.Msg {
							return TabError{Message: "Failed to connect to cluster", Err: err}
						},
					)
				}
				m.StartGeneration()
				if !m.shouldTick {
					m.shouldTick = true
					return m, tea.Tick(time.Second, func(time.Time) tea.Msg {
						return tickMsg{}
					})
				}
				return m, nil
			} else {
				m.running = false
				m.status = "Stopping..."
				m.stopGeneration()
				m.shouldTick = false
				m.status = "Waiting to start"
			}

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
