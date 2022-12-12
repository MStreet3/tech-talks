package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

func workPulse(done <-chan struct{}) (<-chan int, <-chan struct{}) {
	dataCh := make(chan int)
	heartbeatCh := make(chan struct{}, 1)

	go func() {
		defer close(dataCh)
		defer close(heartbeatCh)

		for {
			select {
			case <-done:
				fmt.Println("forwarder: canceled by parent")
				return
			case <-time.After(100 * time.Millisecond):
				// send after artificial delay
				dataCh <- rand.Int()

				// beat for each unit of work
				select {
				case heartbeatCh <- struct{}{}:
				default:
				}
			}
		}
	}()

	return dataCh, heartbeatCh
}

func forwarder(done <-chan struct{}, pulseInterval time.Duration) (<-chan int, <-chan struct{}) {
	dataCh := make(chan int)
	heartbeatCh := make(chan struct{})

	go func() {
		defer close(dataCh)

		for {

			pulse := time.Tick(pulseInterval / 2)

			var sent bool
			for !sent {
				select {
				case <-done:
					fmt.Println("forwarder: canceled by parent")
					return
				case <-pulse:
					select {
					case heartbeatCh <- struct{}{}:
					default:
					}
				case <-time.After(pulseInterval * time.Duration(rand.Float64()+1)):
					// send after artificial delay
					dataCh <- rand.Int()
					sent = true
				}
			}

		}
	}()

	return dataCh, heartbeatCh
}

func reader(done <-chan struct{}, dataCh <-chan int) <-chan struct{} {
	terminated := make(chan struct{})

	go func() {
		defer close(terminated)
		for {
			select {
			case <-done:
				fmt.Println("reader: canceled by parent")
				return
			case num, ok := <-dataCh:
				if !ok {
					fmt.Println("reader: data chan closed, stopping read")
					return
				}

				fmt.Println(num)
			}
		}
	}()

	return terminated
}

func main() {
	var (
		interval     = 100 * time.Millisecond
		ctx, cancel  = context.WithCancel(context.Background())
		dataCh, hbCh = forwarder(ctx.Done(), interval)

		// separate reading from forwarding
		stopReading = make(chan struct{})
		reading     = reader(stopReading, dataCh)
	)

	go func() {
		// Cancel the forwarder after a set number of heart beats.
		defer cancel()

		var count int
		for {
			timeout := time.After(5 * interval)
			select {
			case <-timeout:
				fmt.Println("timed out")
			case <-hbCh:
				fmt.Println("beat")
				if count == 5 {
					close(stopReading)
				}
				count++
			}
		}
	}()

	// wait for cancelations
	<-reading
	<-ctx.Done()
	fmt.Println("parent: all routines dead")
}
