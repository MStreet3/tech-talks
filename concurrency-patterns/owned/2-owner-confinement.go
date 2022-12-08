package owned

import (
	"fmt"
)

// "channel confinement pattern"

// forwarder creates a new channel of integers so that it is only
// accessible within its lexical scope.  The channel is returned immediately
// and a goroutine is launched that writes to the channel and will eventually
// close the channel.  Only one goroutine can ever close the channel.
func forwarder(src <-chan int) <-chan int {
	sink := make(chan int)

	go func() {
		defer close(sink) // <4>

		for val := range src {
			sink <- val
		}
	}()

	return sink // <3>
}

func consumer(src <-chan int) {
	for num := range src {
		fmt.Println(num)
	}
}

func main() {
	var (
		data   = []int{1, 2, 3, 4}
		source = make(chan int, len(data))
	)

	// Load the source channel with data.
	for _, val := range data {
		source <- val
	}

	// Close the source channel after each value from data has been
	// written to it.
	close(source) // <1>

	// The call to the forwarder function is done synchronously in the main
	// goroutine and returns a new channel immediately.
	sink := forwarder(source) // <2>

	consumer(sink) // <5>
}
