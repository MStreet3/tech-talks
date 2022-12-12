# Channel Generators

Channels are one of the core synchronization primities in Go and the backbone
of the Go philosophy on concurrency. Channels enable programs
to share values in such a way that only one goroutine has access to a value at
a time. The Go team slogan underscores the importance of channels:

> Do not communicate by sharing memory; instead, share memory by communicating.

There are three actions that a program can perform on a channel: read, write
or close. A close is a special kind of write that indicates no further values
will ever be written to the channel. Channels can exist in a number of states
and the table below shows the effect that each of these actions will have on
a channel based on its current state.

| Channel State             | Read                       | Write       | Close                                                   |
| ------------------------- | -------------------------- | ----------- | ------------------------------------------------------- |
| `nil`                     | Block                      | Block       | **panic**                                               |
| **closed**                | \<Default Value\>, `false` | **panic**   | **panic**                                               |
| Open, unbuffered          | Block                      | Block       | Closes channel                                          |
| Open, not empty, not full | Value                      | Write Value | Closes channel; reads succeed until channel is drained. |
| Open, empty               | Block                      | Write Value | Closes channel                                          |
| Open and full             | Value                      | Block       | Closes channel; reads succeed until channel is drained. |

_There are three critical channel states to know as taking the wrong action will cause a panic!_

## Adhoc Confinement

From the table above we can see that it is critical for developers to know the
state of channels before invoking either a read, write or close. In fact, developers
must restrict certain actions from ever occuring based on the channel state.

_How does this happen in practice?_

A typical solution to the management of channel state is to apply an _ad hoc_ set
of principles that may or may not be communicated either via comments or informally.
Some of these _ad hoc_ rules may include:

- Use Go typing to specify if a channel is read only or write only
- Never close (or avoid closing) any channel passed as a function parameter
- Wrap a channel close in a _sync.Once.Do()_ call

```go
// forwarder reads values from src channel and writes them on to the sink
// channel.  forwarder closes the sink channel once the src channel is empty.
func forwarder(sink chan<- int, src <-chan int) {
	defer close(sink)

	for val := range src {
		sink <- val
	}
}

// consumer reads values from the src channel and writes them to the standard
// output.  consumer reads indefinitely until the src channel is closed.
func consumer(src <-chan int) {
	for num := range src {
		fmt.Println(num)
	}
}

func main() {
	var (
		data   = []int{1, 2, 3, 4}
		source = make(chan int, len(data)) // <1>
		sink   = make(chan int)
	)

	// Load the source channel with data.
	for _, val := range data {
		source <- val
	}

	// Launch a goroutine to forward the data from the source channel and
	// onto the sink channel.  Adhoc convention dictates that only the
	// forwarder function closes the sink channel.
	go forwarder(sink, source) // <2>

	// Read all the values from the sink channel.  The consumer function requires
	// the adhoc convention that the input channel is eventually closed
	// otherwise it would block forever.
	consumer(sink) // <3>
}
```

### When can this go wrong?

**1** Here we instantiate a buffered source channel that will be filled with the
values from the `data` slice.

**2** The forwarder function is called as a goroutine via the `go` keyword.  
The forwarder function follows the _ad hoc_ convention that it alone will close
the `sink` channel. Any other closes after (or even _before_) the forwarder
closes `sink` will **panic**.

**3** Here we give the consumer function the `sink` channel. The _ad hoc_ convention
is that the `sink` channel is closed somewhere else, which prevents this `main`
goroutine from blocking indefinitely. The consumer function's `src` parameter is
statically typed to be read-only, which prevents the consumer function from closing
the channel.

## Channel Generators and Lexical Confinement

_Static lexical scope_ or _lexical confinement_ is a technique that limits the
usable scope of a variable to the block of code within which it is defined and
these limits are imposed by the compiler at compile time. Lexical confinement
in go is a much more powerful tool for channel confinement than _ad hoc_
confinement because it applies the power of the compiler to enforce proper
code execution. Below is an example that updates the previous `forwarder`
function to rely on lexical confinement:

```go
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

// ... consumer is unchanged

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

	consumer(sink)
}
```

**1** The main goroutine closes the `source` channel after all values from
data are written to it. This close signals to the subsequent channel consumers
that they have read all the enclosed data.

**2** Now the `forwarder` function is called without the `go` keyword because
it returns immediately with a new channel that will eventually close.

**3** `forwarder` creates a new channel of integers so that it is only
accessible within its lexical scope. The channel is returned immediately
and a goroutine is launched that writes to the channel and will eventually
close the channel.

**4** Only one goroutine can ever close the channel.

Writing a function that returns a channel is also referred to as the [generator pattern](https://go.dev/talks/2012/concurrency.slide#25).

Channel generators have many useful qualities.

### Channels as Handlers

Let's first update the `forwarder` function to produce an infinite stream of
data as a service:

```go
func forwarder() <-chan int {
	sink := make(chan int)

	go func() {
		defer close(sink)
		for {
			delay := time.Duration(rand.Intn(1e3)) * time.Millisecond
			<-time.After(delay)
			sink <- rand.Int()
		}
	}()

	return sink
}
```

Next, we can consume the values independently:

```go
func identifier(name string, src <-chan int) <-chan string {
	logs := make(chan string)

	go func() {
		defer close(logs)

		for val := range src {
			logs <- fmt.Sprintf("%s: %d", name, val)
		}
	}()

	return logs
}

func main() {
	sink1 := identifier("sink1", forwarder()) // <1>
	sink2 := identifier("sink2", forwarder())

	// Take five values for each stream.
	for i := 0; i < 5; i++ {
		fmt.Println(<-sink1)
		fmt.Println(<-sink2)
	}
}
```

**1** The identifier function can be chained together with the infinite source
to create a simple pipeline.

## Channel Cleanup

Another important axiom of proper channel usage is:

> If a goroutine is responsible for creating a goroutine, it is also responsible for ensuring that it can stop the goroutine.

```go
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

func main() {
	stop := make(chan struct{}) // <1>
	sink1 := identifier(stop, "sink1", forwarder(stop))
	sink2 := identifier(stop, "sink2", forwarder(stop))

	time.AfterFunc(1*time.Second, func() {
		defer fmt.Println("service shutdown complete")
		defer close(stop) // <2>
		fmt.Println("stopping the service")
	})

	// Take five values for each stream.
	for i := 0; i < 5; i++ {
		fmt.Println(<-sink1)
		fmt.Println(<-sink2)
	}
}
```

**1** Here we create a `stop` channel with the only purpose of broadcasting to all
child goroutines that they must exit. This pattern leverages the feature that a
closed channel can be read from indefinitely and never blocks.

**2** Here we close the `stop` channel after one second, which will prevent the
`for` loop from printing all of its values (at least has a high-probability to
stop the loop before `i == 5`).

**3** Each function now uses a `select` statement to either read from the `stop`
channel or to complete the other communication.

## Fan-in Pattern

In the previous example we are sequentially waiting to receive from `sink1` and
then `sink2`, but the time to produce random values is also random. We are
actually agnostic to which logger reports a value first. To handle each received
log as they arrive we should _fan-in_ the results to a single channel:

```go
func merge[T any](stop <-chan struct{}, chs ...<-chan any) <-chan any {
	merged := make(chan any)

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
	done := make(chan struct{})

	sink1 := identifier(stop, "sink1", forwarder(stop))
	sink2 := identifier(stop, "sink2", forwarder(stop))

	merged := merge(stop, sink1, sink2) // <1>

	time.AfterFunc(5*time.Second, func() {
		defer fmt.Println("service shutdown complete")

		fmt.Println("stopping the service")
		close(stop)
		<-done
	})

	// Read logs until merged is closed.
	for log := range merged {
		fmt.Println(log)
	}
	close(done)
}
```

**1** Here is where we merge the two channels into a single stream of values.
By _fanning-in_ the two handlers we can receive values from the first available
worker.

**2** The select statement of the merge function waits to receive a signal either
from the `stop` channel or one of the input channels. How could this method be
modified to accept a variadic slice of channels?

## Heartbeats

A worker should communicate its liveness and a great option for tracking

### Limitations

Lexical confinement is a great tool for ensuring safe use of channels and for
pipeline development, however, it can in practice have some limitations:

- It is often "simpler" to write code with _ad hoc_ conventions
- It can be difficult to refactor a codebase that uses many _ad hoc_ conventions
  into one that relies on lexical scoping and compile time protections
- Method definitions can grow as multiple closures are defined, which can be
  confusing to developers unfamiliar with lexical scoping
- Use of lexical scoping in practice will often require additional _design
  patterns_ such as the _Adapter_ or _Factory_ patterns
- Practical lexical scoping often necessitates utility functions for dealing
  with channels returned from method and function calls
