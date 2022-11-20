package main

import (
	"context"
	"fmt"
	"time"
)

func forwarder(done <-chan struct{}, data []int) (<-chan int, <-chan struct{}) {
	dataCh := make(chan int)
	heartbeatCh := make(chan struct{}, 1)

	go func() {
		defer close(dataCh)
		defer close(heartbeatCh)

		for i := range data {
			// beat for each unit of work
			select {
			case heartbeatCh <- struct{}{}:
			default:
			}

			select {
			case <-done:
				fmt.Println("forwarder: canceled by parent")
				return
			case <-time.After(100 * time.Millisecond):
				// send after artificial delay
				dataCh <- data[i]
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
		ctx, cancel  = context.WithCancel(context.Background())
		data         = []int{1, 2, 3, 4}
		dataCh, hbCh = forwarder(ctx.Done(), data)
		reading      = reader(ctx.Done(), dataCh)
	)

	// cancel after a set number of beats
	go func() {
		defer cancel()
		for beats := 0; beats < 3; beats++ {
			<-hbCh
		}
	}()

	// wait for cancelations
	<-reading
	fmt.Println("parent: all routines dead")
}
