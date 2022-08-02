package main

import (
	"fmt"
	"context"
	"time"
)

type Data struct {
	Err error
	N int
}

type Service struct {
	stopCh chan struct{}
}

// channel confinement + parental controls
// chanutils.CtxOrDone(ctx, ch)
func forwarder(done <-chan struct{}, data []int) <-chan int {
	dataCh := make(chan int)

	go func(){
		defer close(dataCh)
		for i := range data {
			select{
			case <-done:
				fmt.Println("forwarder: canceled by parent")
				return
			case <-time.After(500 * time.Millisecond):

				// send after artificial delay
				dataCh <- data[i]
			}
		}
	}()

	return dataCh
}

func reader(done <-chan struct{}, dataCh <-chan int) {
	for {
		select{
			case <-done:
				fmt.Println("reader: canceled by parent")
				return
			case num, ok := <-dataCh:
				if !ok{
					fmt.Println("reader: data chan closed, stopping read")
					return
				}

				fmt.Println(num)
		}
	}
}


func main(){
	var done bool
	ctx, cancel := context.WithCancel(context.Background())
	data := []int{1, 2, 3, 4}

	dataCh := forwarder(ctx.Done(), data)

	time.AfterFunc(1500 * time.Millisecond, func(){
		cancel()
	})

	reader(ctx.Done(), dataCh)
	
	// wait for cancelations
	for !done {
		if _, ok := <-dataCh; !ok {
			select{
			case <-ctx.Done():
				done = true
			}
		}
	}
	fmt.Println("parent: all routines dead")
}