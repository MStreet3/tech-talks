package owned

import (
	"conpatterns/entities"
	"conpatterns/errs"
	"sync"
	"sync/atomic"
)

// The StartStopOnceAdapter object can be started and stopped only once.
// Repeated calls to Start and Stop will return errors.

// StartStopOnceAdapter uses two atomic.Bool properties to track the state of
// either started or stopped.  This avoids the need for a constructor because
// nil values of started and stopped are valid, unlike a nil channel which
// cannot be closed.
type StartStopOnceAdapter struct {
	mu        sync.RWMutex
	impl      entities.StartStopBuilder
	stop      func() <-chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
	started   atomic.Bool
	stopped   atomic.Bool
}

var _ entities.StartStopper = (*StartStopOnceAdapter)(nil)

func (s *StartStopOnceAdapter) Start() error {
	if s.IsStarted() {
		return errs.ErrAlreadyStarted
	}

	if s.IsStopped() {
		return errs.ErrAlreadyStopped
	}

	s.startOnce.Do(
		func() {
			s.mu.Lock()
			defer s.mu.Unlock()

			start, stop := s.impl.Build()
			s.stop = stop
			started := start()
			<-started
			s.started.Store(true)
		})
	return nil
}

func (s *StartStopOnceAdapter) Stop() error {
	if s.IsStopped() {
		return errs.ErrAlreadyStopped
	}

	if !s.IsStarted() {
		return errs.ErrNotYetStarted
	}

	s.stopOnce.Do(
		func() {
			s.mu.Lock()
			defer s.mu.Unlock()

			done := s.stop()
			<-done
			s.stopped.Store(true)
		},
	)
	return nil
}

func (s *StartStopOnceAdapter) IsStarted() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state := s.started.Load()
	return state
}

func (s *StartStopOnceAdapter) IsStopped() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state := s.stopped.Load()
	return state
}
