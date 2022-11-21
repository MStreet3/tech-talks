package errs

import "errors"

var (
	ErrAlreadyStopped = errors.New("instance is already stopped")
	ErrAlreadyStarted = errors.New("instance is already started")
	ErrNotYetStarted  = errors.New("instance is not yet started")
)
