package errors

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type ErrorTimeoutMsg struct{}

type ErrorOverlay struct {
	Message   string
	StartTime time.Time
	Duration  time.Duration
	ShowTime  time.Duration
}

type ShowErrorMsg struct {
	Message string
}

func NewErrorOverlay(message string) *ErrorOverlay {
	return &ErrorOverlay{
		Message:   message,
		StartTime: time.Now(),
		ShowTime:  3 * time.Second,
	}
}

func ErrorTimeout() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return ErrorTimeoutMsg{}
	})
}
