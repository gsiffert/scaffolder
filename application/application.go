package application

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Vorian-Atreides/scaffolder"
)

type Application struct {
	name           string
	version        string
	gracefulPeriod time.Duration

	inventory  *scaffolder.Inventory
	components []scaffolder.Component
}

func (a *Application) Default() {
	a.name = os.Args[0]
	a.version = "0.0.0"
	a.gracefulPeriod = time.Second
	a.inventory = scaffolder.New()
}

func New(opts ...scaffolder.Option) (*Application, error) {
	app := &Application{}
	return app, scaffolder.Init(app, opts...)
}

func WithGracefulPeriod(duration time.Duration) scaffolder.Option {
	return func(a *Application) error {
		a.gracefulPeriod = duration
		return nil
	}
}

func WithComponent(component scaffolder.Component, opts ...scaffolder.Option) scaffolder.Option {
	return func(a *Application) error {
		a.components = append(a.components, component)
		a.inventory.Add(component, opts...)
		return nil
	}
}

func WithVersion(version string) scaffolder.Option {
	return func(a *Application) error {
		a.version = version
		return nil
	}
}

func WithName(name string) scaffolder.Option {
	return func(a *Application) error {
		a.name = name
		return nil
	}
}

func (a *Application) validate() error {
	for _, component := range a.components {
		// Validate the components before starting them.
		if validator, ok := component.(Validator); ok {
			if err := validator.Validate(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Application) stopWithTimeout(ctx context.Context, s StopHook) func() error {
	return func() error {
		ctx, cancel := context.WithTimeout(ctx, a.gracefulPeriod)
		defer cancel()
		return s.Stop(ctx)
	}
}

// Run will start by linking the components with the scaffolder Inventory.
//
// Then every components implementing the Validator interface will be validated,
// the application will abort if at least one error has been returned.
//
// Otherwise, the application will start the components implementing the StartHook interface
// in a separated Goroutine. Returning at least one error from the Starts will stop the Application.
//
// Finally before the application terminate, it will notify the components implementing
// the StopHook interface to allow them to gracefully shutdown.
func (a *Application) Run(ctx context.Context) (err error) {
	if err := a.inventory.Compile(); err != nil {
		return err
	}
	if err := a.validate(); err != nil {
		return err
	}

	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, syscall.SIGINT, syscall.SIGTERM)

	wg := sync.WaitGroup{}
	defer wg.Wait()

	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	runtimeErr := make(chan error)
	for _, component := range a.components {
		// Start the component in its own Goroutine and
		// block until the scheduler started it.
		if s, ok := component.(StartHook); ok {
			cond := make(chan struct{})
			go func(s StartHook) {
				close(cond)
				select {
				case <-ctx.Done():
				case runtimeErr <- s.Start(childCtx):
				}
			}(s)
			select {
			case <-ctx.Done():
				return
			case <-signalC:
				return
			case <-cond:
			}
		}

		// Hook the call to stop the component when shutting down the application,
		// we stop them sequentially.
		if s, ok := component.(StopHook); ok {
			wg.Add(1)
			stopper := a.stopWithTimeout(childCtx, s)
			defer func() {
				if stopErr := stopper(); stopErr != nil && err == nil {
					err = stopErr
				}
				wg.Done()
			}()
		}
	}

	select {
	case <-ctx.Done():
	case <-signalC:
	case err = <-runtimeErr:
	}
	return err
}

type Validator interface {
	Validate() error
}

type StartHook interface {
	Start(context.Context) error
}

type StopHook interface {
	Stop(context.Context) error
}
