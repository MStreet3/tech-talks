package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

var ErrRetryJob = errors.New("retry me")

type Job struct {
	ID string
}

func NewJob(n int) Job {
	return Job{
		ID: fmt.Sprintf("id-%d", n),
	}
}

func (j Job) Run() error {
	return ErrRetryJob
}

type Batch struct {
	jobs chan Job
	cap  int
	mu   sync.Mutex
}

func NewBatch(n int) *Batch {
	return &Batch{
		cap:  n,
		jobs: make(chan Job, 1),
	}
}

func (b *Batch) Add(j Job) {
	select {
	case b.jobs <- j:
	default:
		// Drop the job if the queue is full.
	}
}

func (b *Batch) Start() {
	go batchJobs(b.jobs, b.cap)
}

func batchJobs(ch chan Job, cap int) {
	var batch []Job
	for job := range ch {
		batch = append(batch, job)
		if len(batch) == cap {
			log.Println("calling process")
			go processJobs(batch, ch)
			batch = nil
		}
	}
}

func processJobs(batch []Job, ch chan<- Job) {
	for _, job := range batch {
		err := job.Run()
		if IsRetryableError(err) {
			// add the job back to the work pool
			log.Printf("putting job back on queue: %s\n", job.ID)
			ch <- job
		} else if err != nil {
			log.Printf("job failed, discarding: %s\n", job.ID)
		}
	}
}

func IsRetryableError(e error) bool {
	return errors.Is(e, ErrRetryJob)
}

// the original implementation can easily deadlock because processJobs is blocking on write
// improvement 1: async call to processJobs
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	batch := NewBatch(5)
	batch.Start()

	go func() {
		for i := 0; i < 6; i++ {
			select {
			case <-ctx.Done():
				return
			default:
				batch.Add(NewJob(i))
			}
		}
	}()

	<-ctx.Done()
}
