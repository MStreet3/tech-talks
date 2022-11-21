package owned

import (
	"fmt"
)

// "channel confinement pattern"
// forwarder confines a new channel of integers so that it is only
// accessible (i.e., closeable) within its lexical scope
func forwarder(data []int) <-chan int {
	dataCh := make(chan int)

	go func() {
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

func main() {
	data := []int{1, 2, 3, 4}

	handleData := forwarder(data)

	reader(handleData)
}
