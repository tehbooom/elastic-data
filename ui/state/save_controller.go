package state

import (
	"time"
)

// SaveController manages when to save the application state
type SaveController struct {
	appState       *AppState
	debounceTimer  *time.Timer
	debouncePeriod time.Duration
}

// NewSaveController creates a new save controller
func NewSaveController(state *AppState) *SaveController {
	return &SaveController{
		appState:       state,
		debouncePeriod: 2 * time.Second,
	}
}

// MarkDirty marks the state as dirty and schedules a save
func (s *SaveController) MarkDirty() {
	s.appState.Dirty = true

	// Cancel existing timer if there is one
	if s.debounceTimer != nil {
		s.debounceTimer.Stop()
	}

	// Schedule a new save
	s.debounceTimer = time.AfterFunc(s.debouncePeriod, func() {
		if s.appState.Dirty {
			s.appState.SaveIntegrations()
		}
	})
}

// SaveNow forces an immediate save if there are changes
func (s *SaveController) SaveNow() {
	if s.debounceTimer != nil {
		s.debounceTimer.Stop()
		s.debounceTimer = nil
	}

	if s.appState.Dirty {
		s.appState.SaveIntegrations()
	}
}
