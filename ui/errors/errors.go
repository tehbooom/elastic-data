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
	Fatal     bool
}

type ShowErrorMsg struct {
	Message string
	Fatal   bool
}

func NewErrorOverlay(message string, fatal bool) *ErrorOverlay {
	return &ErrorOverlay{
		Message:   message,
		StartTime: time.Now(),
		ShowTime:  5 * time.Second,
		Fatal:     fatal,
	}
}

func ErrorTimeout() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return ErrorTimeoutMsg{}
	})
}
