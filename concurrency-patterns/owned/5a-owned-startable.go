package owned

import (
	"sync"

	"conpatterns/entities"
)

var _ entities.StartStopperFactory = (*ownedStartable)(nil)

type ownedStartable struct {
}

func (owned *ownedStartable) Build() (start func() <-chan struct{}, stop func() <-chan struct{}) {
	var (
		wg        sync.WaitGroup
		startOnce sync.Once
		stopOnce  sync.Once

		quit    = make(chan struct{})
		started = make(chan struct{})
		stopped = make(chan struct{})
	)

	start = func() <-chan struct{} {
		// Create a goroutine to initialize startup.  Startup is only
		// initialized once.
		go func() {
			defer close(started)

			startOnce.Do(func() {
				isRunning := owned.loop(quit)

				wg.Add(1)
				go func() {
					defer wg.Done()

					<-isRunning
				}()
			})
		}()

		return started
	}

	stop = func() <-chan struct{} {
		// Create a goroutine to initiate shutdown.  Shutdown can only
		// be initiated once.
		go func() {
			defer close(stopped)

			stopOnce.Do(func() {
				defer wg.Wait()

				close(quit)
			})
		}()

		return stopped
	}

	return
}

func (owned *ownedStartable) loop(stop <-chan struct{}) <-chan struct{} {
	return nil
}
