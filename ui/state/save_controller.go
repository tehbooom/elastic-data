package state

import (
	"time"

	"github.com/charmbracelet/log"
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
	log.Debug("App state marked dirty, scheduling save")

	// Cancel existing timer if there is one
	if s.debounceTimer != nil {
		s.debounceTimer.Stop()
	}

	// Schedule a new save
	s.debounceTimer = time.AfterFunc(s.debouncePeriod, func() {
		if s.appState.Dirty {
			log.Debug("Debounce timer triggered, saving app state")
			s.appState.SaveIntegrations()
		}
	})
}

// SaveNow forces an immediate save if there are changes
func (s *SaveController) SaveNow() {
	log.Debug("SaveNowCalled")
	if s.debounceTimer != nil {
		s.debounceTimer.Stop()
		s.debounceTimer = nil
	}

	if s.appState.Dirty {
		log.Debug("App state is dirty, saving now")
		s.appState.SaveIntegrations()
	} else {
		log.Debug("App state is not dirty, nothing to save")
	}
}
