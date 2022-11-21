package adhoc

import (
	"conpatterns/entities"
	"conpatterns/errs"
	"sync"
)

var _ entities.StartStopper = (*adhocStarter)(nil)

type adhocStarter struct {
	startOnce sync.Once
	stopOnce  sync.Once
	wg        sync.WaitGroup
	quit      chan struct{}
	started   chan struct{}
}

func NewAdhocStartable() *adhocStarter {
	return &adhocStarter{
		quit:    make(chan struct{}),
		started: make(chan struct{}),
	}
}

func (adhoc *adhocStarter) Start() error {
	if adhoc.IsStarted() {
		return errs.ErrAlreadyStarted
	}

	if adhoc.IsStopped() {
		return errs.ErrAlreadyStopped
	}

	adhoc.startOnce.Do(adhoc.start)
	return nil
}

func (adhoc *adhocStarter) start() {
	defer close(adhoc.started)

	isRunning := adhoc.loop(adhoc.quit)

	adhoc.wg.Add(1)
	go func() {
		defer adhoc.wg.Done()

		<-isRunning
	}()
}

func (adhoc *adhocStarter) Stop() error {
	if adhoc.IsStopped() {
		return errs.ErrAlreadyStopped
	}

	if !adhoc.IsStarted() {
		return errs.ErrNotYetStarted
	}

	adhoc.stopOnce.Do(adhoc.stop)
	return nil
}

func (adhoc *adhocStarter) stop() {
	defer adhoc.wg.Wait()

	close(adhoc.quit)
}

func (adhoc *adhocStarter) IsStarted() bool {
	select {
	case <-adhoc.started:
		return true
	default:
	}
	return false
}

func (adhoc *adhocStarter) IsStopped() bool {
	select {
	case <-adhoc.quit:
		return true
	default:
	}
	return false
}

func (adhoc *adhocStarter) loop(stop <-chan struct{}) <-chan struct{} {
	return nil
}
