package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func forwarder(stop <-chan struct{}) <-chan int {
	sink := make(chan int)

	go func() {
		defer close(sink)
		for {
			delay := time.Duration(rand.Intn(1e3)) * time.Millisecond
			select {
			case <-stop: // <3>
				return
			case <-time.After(delay):
				sink <- rand.Int()
			}
		}
	}()

	return sink
}

func identifier(stop <-chan struct{}, name string, src <-chan int) <-chan string {
	logs := make(chan string)

	go func() {
		defer close(logs)

		for val := range src {
			select {
			case <-stop: // <3>
				return
			case logs <- fmt.Sprintf("%s: %d", name, val):
			}
		}
	}()

	return logs
}

func consumer(stop <-chan struct{}, src <-chan string) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)

		for val := range src {
			select {
			case <-stop: // <3>
				return
			default:
				fmt.Println(val)
			}
		}
	}()

	return done
}

func merge[T any](stop <-chan struct{}, chs ...<-chan T) <-chan T {
	merged := make(chan T)

	var wg sync.WaitGroup

	wg.Add(len(chs))
	for _, ch := range chs {
		ch := ch

		go func() {
			defer wg.Done()
			for val := range ch {
				select {
				case merged <- val:
				case <-stop:
					return
				}
			}
		}()
	}

	go func() {
		defer close(merged)

		wg.Wait()
	}()

	return merged
}

func main() {
	stop := make(chan struct{})
	shutdown := make(chan struct{})

	// Start forwarding messages.
	sink1 := identifier(stop, "sink1", forwarder(stop))
	sink2 := identifier(stop, "sink2", forwarder(stop))

	// Merge all messages onto a single channel to read from.
	merged := merge(stop, sink1, sink2) // <1>

	isReading := consumer(stop, merged)

	time.AfterFunc(2*time.Second, func() {
		defer fmt.Println("service shutdown complete")
		defer close(shutdown)

		// Signal to stop all child routines and wait
		// until done reading.
		fmt.Println("stopping the service")
		close(stop)
		<-isReading
	})

	// Await shutdown signal once done reading.
	<-shutdown
}
