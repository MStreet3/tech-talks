package owned

import (
	"conpatterns/entities"
	"conpatterns/errs"
)

// An OwnedStarter implementation can be adapted to meet the StartStopper
// interface.

// The MultiStartAdapter object can be started and stopped repeatedly.  Each
// call to start will initialize a new goroutine of the implementation.  Calls
// to stop will cleanup the goroutine.
type MultiStartAdapter struct {
	impl entities.OwnedStarter
	stop func() <-chan struct{}
}

var _ entities.StartStopper = (*MultiStartAdapter)(nil)

func (m *MultiStartAdapter) Start() error {
	start, stop := m.impl.Build()
	m.stop = stop
	start()
	return nil
}

func (m *MultiStartAdapter) Stop() error {
	if m.stop == nil {
		return errs.ErrNotYetStarted
	}
	done := m.stop()
	<-done
	return nil
}
