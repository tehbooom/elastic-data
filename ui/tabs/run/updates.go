package run

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/tehbooom/elastic-data/ui/errors"
)

type TickMsg struct{}

func (m *TabModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		if !m.programContext.IsRunning() {
			return m, nil
		}

		return m, tea.Tick(time.Second, func(time.Time) tea.Msg {
			return TickMsg{}
		})

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			if m.programContext.IsRunning() {
				m.programContext.SetRunning(false)
				m.status = "Stopping..."
				m.stopGeneration()
				m.status = "Waiting to start"
			}
			return m, nil
		case "enter":
			if !m.programContext.IsRunning() {
				m.status = StartedMsg
				err := m.programContext.ESClient.TestConnection()
				if err != nil {
					log.Debug(err)
					return m, func() tea.Msg {
						return errors.ShowErrorMsg{Message: fmt.Sprintf("Error: %v", err)}
					}
				}

				err = m.programContext.KBClient.TestConnection()
				if err != nil {
					log.Debug(err)
					return m, func() tea.Msg {
						return errors.ShowErrorMsg{Message: fmt.Sprintf("Error: %v", err)}
					}
				}

				err = m.StartGeneration()
				if err != nil {
					log.Debug(err)
					return m, func() tea.Msg {
						return errors.ShowErrorMsg{Message: fmt.Sprintf("Error generating data: %v", err)}
					}
				}

				m.programContext.SetRunning(true)
				return m, tea.Tick(time.Second, func(time.Time) tea.Msg {
					return TickMsg{}
				})
			} else {
				m.programContext.SetRunning(false)
				m.status = "Stopping..."
				m.stopGeneration()
				m.status = "Waiting to start"
				return m, nil
			}
		}
	}

	return m, nil
}
