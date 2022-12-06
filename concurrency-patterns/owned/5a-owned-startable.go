package owned

import (
	"sync"

	"conpatterns/entities"
)

var _ entities.StartStopBuilder = (*OwnedBuilder)(nil)

type OwnedBuilder struct {
}

func (owned *OwnedBuilder) Build() (start func() <-chan struct{}, stop func() <-chan struct{}) {
	var (
		// The startOnce and stopOnce primitives ensure that subsequent
		// calls to either the start or stop functions will not execute
		// any actions other than immediately closing their returned
		// channels.  A new call to Build is required to generate new
		// functions to initialize and shutdown the loop goroutine.
		startOnce sync.Once
		stopOnce  sync.Once

		// The wait group is owned by the Build method and synchronizes
		// the start and stop functions.  All goroutines initiated
		// in the start function are added to the waitgroup.  The stop
		// function knows that shutdown is complete once the waitgroup
		// counter is emptied.
		wg sync.WaitGroup

		// The quit channel is owned by the Build method yet safely
		// shared across the start and stop functions.  All child
		// goroutines created in the start function should accept the
		// quit channel as a parameter and clean themselves up (i.e.,
		// shutdown) once the close signal is received on quit.
		quit = make(chan struct{})
	)

	start = func() <-chan struct{} {
		started := make(chan struct{})
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
		stopped := make(chan struct{})
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

func (owned *OwnedBuilder) loop(stop <-chan struct{}) <-chan struct{} {
	return nil
}
