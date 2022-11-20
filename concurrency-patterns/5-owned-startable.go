package main

import "sync"

type ownedStartable struct {
	startOnce sync.Once
	mu        sync.Mutex
	stopCh    chan struct{}
}

func (owned *ownedStartable) Start() (func() (<-chan struct{}, error), error) {
	var (
		wg     sync.WaitGroup
		stopCh = owned.getStopCh()

		start = func() {
			quit := make(chan struct{})
			isRunning := owned.loop(quit)

			wg.Add(1)
			go func() {
				defer wg.Done()
				defer close(quit)

				<-stopCh
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()

				<-isRunning
			}()
		}

		stop = func() (<-chan struct{}, error) {
			done := make(chan struct{})

			go func() {
				defer close(done)
				select {
				case stopCh <- struct{}{}:
					wg.Wait()
					owned.mu.Lock()
					owned.startOnce = sync.Once{}
					owned.mu.Unlock()
				default:
					// no routine is reading or the signal
					// has already been sent
				}
			}()

			return done, nil
		}
	)

	owned.startOnce.Do(start)

	return stop, nil
}

func (owned *ownedStartable) getStopCh() chan struct{} {
	if owned.stopCh == nil {
		owned.stopCh = make(chan struct{}, 1)
	}
	return owned.stopCh
}

func (owned *ownedStartable) loop(stop <-chan struct{}) <-chan struct{} {
	return nil
}
