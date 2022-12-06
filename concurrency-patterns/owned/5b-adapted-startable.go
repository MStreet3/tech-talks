package owned

import (
	"conpatterns/entities"
	"conpatterns/errs"
	"sync"
)

// An OwnedStarter implementation can be adapted to meet the StartStopper
// interface.

// The MultiStartAdapter object can be started and stopped repeatedly.  Each
// call to start will initialize a new goroutine of the implementation.  Calls
// to stop will cleanup the goroutine.
type MultiStartAdapter struct {
	mu   sync.Mutex
	impl entities.StartStopperFactory
	stop func() <-chan struct{}
}

var _ entities.StartStopper = (*MultiStartAdapter)(nil)

func (m *MultiStartAdapter) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.stop != nil {
		return errs.ErrAlreadyStarted
	}

	start, stop := m.impl.Build()
	m.stop = stop
	started := start()
	<-started
	return nil
}

func (m *MultiStartAdapter) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.stop == nil {
		return errs.ErrNotYetStarted
	}

	stopped := m.stop()
	<-stopped
	m.stop = nil
	return nil
}
