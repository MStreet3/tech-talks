package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

func forwarder(done <-chan struct{}, pulseInterval time.Duration) (<-chan int, <-chan struct{}) {
	dataCh := make(chan int)
	heartbeatCh := make(chan struct{}, 1)

	go func() {
		defer close(dataCh)

		for {

			pulse := time.Tick(pulseInterval)

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
				case <-time.After(time.Duration(rand.Intn(100)) * time.Millisecond):
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
		stopReading  = make(chan struct{})
		reading      = reader(stopReading, dataCh)
	)

	// cancel after a set number of beats
	go func() {
		defer cancel()

		var count int
		for {
			timeout := time.After(2 * interval)
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
