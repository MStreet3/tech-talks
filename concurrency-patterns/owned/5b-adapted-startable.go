package owned

import (
	"conpatterns/entities"
	"conpatterns/errs"
	"sync"
)

// A StartStopBuilder implementation can be adapted to meet the StartStopper
// interface.

// The MultiStartAdapter object can be started and stopped repeatedly.  Each
// call to start will initialize a new goroutine of the implementation.  Calls
// to stop will cleanup the goroutine.
type MultiStartAdapter struct {
	mu   sync.Mutex
	impl entities.StartStopBuilder
	stop func() <-chan struct{}
}

var _ entities.StartStopper = (*MultiStartAdapter)(nil)

func (m *MultiStartAdapter) Start() error {
	// Obtain and hold the adapter's lock to safely access the stop
	// function.
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.stop != nil {
		return errs.ErrAlreadyStarted
	}

	start, stop := m.impl.Build()
	started := start()
	<-started

	// Set the stop function on the adapter to indicate that it has been
	// started.
	m.stop = stop
	return nil
}

func (m *MultiStartAdapter) Stop() error {
	// Obtain and hold the adapter's lock to safely access the stop
	// function.
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.stop == nil {
		return errs.ErrNotYetStarted
	}

	stopped := m.stop()
	<-stopped

	// Unset the stop function on the adapter to indicate that it can
	// be started again.
	m.stop = nil
	return nil
}
