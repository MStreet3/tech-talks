package entities

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

type StartStopBuilder interface {
	Build() (start func() <-chan struct{}, stop func() <-chan struct{})
}
