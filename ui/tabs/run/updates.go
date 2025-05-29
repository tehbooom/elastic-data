package run

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
)

type tickMsg struct{}

type refreshMsg struct{}

func (m *TabModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Debug(fmt.Sprintf("TabModel.Update received: %T", msg))
	switch msg := msg.(type) {
	case tickMsg:
		// Only continue ticking if we should be ticking
		if !m.shouldTick {
			log.Debug("Ticker stopped - shouldTick is false")
			return m, nil
		}
		log.Debug("Tick - refreshing view")
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
					log.Debug("Starting ticker for view updates")
					return m, tea.Tick(time.Second, func(time.Time) tea.Msg {
						return tickMsg{}
					})
				}
				log.Debug("Ticker already running - not starting another")
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
