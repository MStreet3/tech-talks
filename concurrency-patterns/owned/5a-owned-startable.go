package owned

import (
	"sync"

	"conpatterns/entities"
)

var _ entities.OwnedStarter = (*ownedStartable)(nil)

type ownedStartable struct {
}

func (owned *ownedStartable) Build() (start func(), stop func() <-chan struct{}) {
	var (
		wg        sync.WaitGroup
		startOnce sync.Once
		stopOnce  sync.Once

		quit = make(chan struct{})
	)

	start = func() {
		startOnce.Do(func() {
			isRunning := owned.loop(quit)

			wg.Add(1)
			go func() {
				defer wg.Done()

				<-isRunning
			}()
		})
	}

	stop = func() <-chan struct{} {
		done := make(chan struct{})

		go func() {
			defer close(done)

			stopOnce.Do(func() {
				defer wg.Wait()

				close(quit)
			})
		}()

		return done
	}

	return
}

func (owned *ownedStartable) loop(stop <-chan struct{}) <-chan struct{} {
	return nil
}
