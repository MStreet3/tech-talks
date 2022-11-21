package owned

import (
	"conpatterns/entities"
	"conpatterns/errs"
	"sync"
)

type StartStopOnceAdapter struct {
	impl      entities.OwnedStarter
	stop      func() <-chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
}

var _ entities.StartStopper = (*StartStopOnceAdapter)(nil)

func (s *StartStopOnceAdapter) Start() error {
	s.startOnce.Do(
		func() {
			start, stop := s.impl.Build()
			s.stop = stop
			start()
		})
	return nil
}

func (s *StartStopOnceAdapter) Stop() error {
	if s.stop == nil {
		return errs.ErrNotYetStarted
	}
	s.stopOnce.Do(
		func() {
			done := s.stop()
			<-done
		},
	)
	return nil
}
