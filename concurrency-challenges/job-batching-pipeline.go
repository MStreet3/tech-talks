package main

import (
	"context"
	"log"
)

func jobPipeline(ctx context.Context, jobQueue []Job) (<-chan struct{}, <-chan error) {
	var (
		ctxwc, cancel = context.WithCancel(ctx)
		n             = 5
		done          = make(chan struct{})
		errCh         = make(chan error)
	)

	jqCh := enqueueArr(ctxwc, jobQueue)
	retries, errors := processor(ctxwc, batch(ctxwc, jqCh, n))
	queuingJobs := enqueuer(ctxwc, retries)
	queuingErrs := enqueuer(ctxwc, errors)

	go func() {
		defer close(done)
		defer close(errCh)
		defer cancel()
		<-queuingJobs
		<-queuingErrs
	}()

	return done, errCh
}

func batch(ctx context.Context, jobs <-chan Job, n int) <-chan []Job {
	batches := make(chan []Job)
	go func() {
		defer close(batches)
		batch := make([]Job, 0)
		for j := range jobs {
			if len(batch) == n {
				select {
				case <-ctx.Done():
					return
				case batches <- batch:
					batch = make([]Job, 0)
				}
				continue
			}
			batch = append(batch, j)
		}
	}()
	return batches
}

func processor(ctx context.Context, batches <-chan []Job) (<-chan Job, <-chan error) {
	var (
		retryable = make(chan Job)
		errCh     = make(chan error)
	)

	go func() {
		defer close(retryable)
		defer close(errCh)
		for {
			select {
			case <-ctx.Done():
				return
			case batch, open := <-batches:
				if !open {
					return
				}
				for _, j := range batch {
					if err := runJob(j); err != nil {
						if !IsRetryableError(err) {
							log.Printf("job failed, discarding: %s\n", j.ID)
							select {
							case <-ctx.Done():
								return
							case errCh <- err:
								continue
							}
						}
						select {
						case <-ctx.Done():
							return
						case retryable <- j:
						}
					}
				}
			}
		}
	}()

	return retryable, errCh
}

func enqueuer[T interface{}](ctx context.Context, src <-chan T) <-chan T {
	dest := make(chan T)
	go func() {
		defer close(dest)
		for j := range src {
			select {
			case <-ctx.Done():
				return
			case dest <- j:
			}
		}
	}()
	return dest
}

func enqueueArr[T interface{}](ctx context.Context, src []T) <-chan T {
	dest := make(chan T)
	go func() {
		defer close(dest)
		for _, j := range src {
			select {
			case <-ctx.Done():
				return
			case dest <- j:
			}
		}
	}()
	return dest
}
