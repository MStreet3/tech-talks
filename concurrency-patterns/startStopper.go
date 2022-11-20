package main

import "errors"

type Starter interface {
	Start() error
}

type Stopper interface {
	Stop() error
}

type StartStopper interface {
	Starter
	Stopper
}

type OwnedStarter interface {
	Start() (func() (<-chan struct{}, error), error)
}

type adapter struct {
	impl OwnedStarter
	stop func() (<-chan struct{}, error)
}

var _ StartStopper = (*adapter)(nil)

func (a *adapter) Start() error {
	fn, err := a.impl.Start()
	if err != nil {
		return err
	}

	a.stop = fn
	return nil
}

func (a *adapter) Stop() error {
	if a.stop == nil {
		return errors.New("method not implemented")
	}
	done, err := a.stop()
	if err != nil {
		return err
	}

	<-done
	return nil
}
