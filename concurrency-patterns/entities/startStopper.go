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

type StartStopperFactory interface {
	Build() (start func() <-chan struct{}, stop func() <-chan struct{})
}
