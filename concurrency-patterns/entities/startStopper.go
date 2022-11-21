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

type OwnedStarter interface {
	Build() (start func(), stop func() <-chan struct{})
}
