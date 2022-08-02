package main

import (
	"fmt"
)

// "channel confinement pattern"
func forwarder(data []int) <-chan int {
	dataCh := make(chan int)

	go func(){
		defer close(dataCh)
		for i := range data {
			dataCh <- data[i]
		}
	}()

	return dataCh
}

func reader(dataCh <-chan int) {
	for num := range dataCh {
		fmt.Println(num)
	}
}

func main(){
	data := []int{1, 2, 3, 4}

	handleData := forwarder(data)

	reader(handleData)
}