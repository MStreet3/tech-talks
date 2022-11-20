package main

import (
	"fmt"
)

// "ad-hoc confinement" -> panics on channels are controlled by a set of ad-hoc rules or conventions in
// the repository

// close a closed chanel -> panic
func forwarder(dataCh chan<- int, data []int) {
	defer close(dataCh)
	for i := range data {
		dataCh <- data[i]
	}
}

func reader(dataCh <-chan int) {
	for num := range dataCh {
		fmt.Println(num)
	}
}

func main() {
	data := []int{1, 2, 3, 4}
	handleData := make(chan int)

	// async call is in main function and not confined to the forwarder
	go forwarder(handleData, data)

	reader(handleData)
}
