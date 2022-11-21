package main

import (
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
