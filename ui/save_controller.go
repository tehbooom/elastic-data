// ui/save_controller.go
package ui

import (
	"log"
	"time"
)

// SaveController manages when to save the application state
type SaveController struct {
	appState       *AppState
	debounceTimer  *time.Timer
	debounceperiod time.Duration
}

// NewSaveController creates a new save controller
func NewSaveController(state *AppState) *SaveController {
	return &SaveController{
		appState:       state,
		debounceperiod: 2 * time.Second,
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
	s.debounceTimer = time.AfterFunc(s.debounceperiod, func() {
		if s.appState.Dirty {
			if err := s.appState.SaveToConfig(); err != nil {
				log.Printf("Error saving config: %v", err)
			}
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
		if err := s.appState.SaveToConfig(); err != nil {
			log.Printf("Error saving config: %v", err)
		}
	}
}
