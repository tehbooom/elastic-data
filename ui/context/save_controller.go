package context

import (
	"time"

	"github.com/charmbracelet/log"
)

type SaveController struct {
	programContext *ProgramContext
	debounceTimer  *time.Timer
	debouncePeriod time.Duration
}

func NewSaveController(state *ProgramContext) *SaveController {
	return &SaveController{
		programContext: state,
		debouncePeriod: 2 * time.Second,
	}
}

func (s *SaveController) MarkDirty() {
	s.programContext.Dirty = true
	log.Debug("App state marked dirty, scheduling save")

	if s.debounceTimer != nil {
		s.debounceTimer.Stop()
	}

	s.debounceTimer = time.AfterFunc(s.debouncePeriod, func() {
		if s.programContext.Dirty {
			log.Debug("Debounce timer triggered, saving")
			s.programContext.SaveIntegrations()
		}
	})
}

func (s *SaveController) SaveNow() {
	log.Debug("SaveNowCalled")
	if s.debounceTimer != nil {
		s.debounceTimer.Stop()
		s.debounceTimer = nil
	}

	if s.programContext.Dirty {
		log.Debug("App state is dirty, saving now")
		s.programContext.SaveIntegrations()
	} else {
		log.Debug("App state is not dirty, nothing to save")
	}
}
