package main

import (
	"context"
	"math/rand"
	"time"
)

func source(stop <-chan struct{}) <-chan int {
	data := make(chan int)

	go func() {
		defer close(data)
		for {
			delay := time.Duration(1000*rand.Float64()) * time.Millisecond
			select {
			case <-stop:
				return
			case <-time.After(delay):
				data <- rand.Int()
			}
		}
	}()

	return data
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	data := source(ctx.Done())
	for range data {

	}
}
