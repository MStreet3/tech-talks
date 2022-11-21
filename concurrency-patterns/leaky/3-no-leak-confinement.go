package leaky

import (
	"context"
	"fmt"
	"time"
)

// channel confinement + parental control
func forwarder(done <-chan struct{}, data []int) <-chan int {
	dataCh := make(chan int)

	go func() {
		defer close(dataCh)
		for i := range data {
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

	return dataCh
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
		ctx, cancel = context.WithCancel(context.Background())
		data        = []int{1, 2, 3, 4}
		dataCh      = forwarder(ctx.Done(), data)
		reading     = reader(ctx.Done(), dataCh)
	)

	time.AfterFunc(300*time.Millisecond, func() {
		cancel()
	})

	// wait for cancelations
	<-reading
	fmt.Println("parent: all routines dead")
}
