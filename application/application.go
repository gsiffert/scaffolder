/*
Package application define a simple application life cycle which should be
robust enough for most of the application or be extended by your own custom
implementation if it does not.

The application will start by initializing every given components, linking them to
each other with the Inventory.

The components which implements the optional Validator interface will then be validated,
any returned error will abort the application.

If no error has been returned, the application will then move to the next phase.
Every components which implements the option StartHook interface will be started
in the same order than it was given in the call to WithComponents.
Returning an error from the Start callback will abort the application.

Finally, the application will run until it receives an interruption signal, or its context
has been canceled or expired, or an error has been returned from the Start callback,

Once the application initiate its interruption, the components implementing
the StopHook interface will be asked to stop and forcefully stopped if they do not perform
after the configured grace period.
The components will be stopped in the reverse order than the one given in the call
to WithComponents.
*/
package application

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Vorian-Atreides/scaffolder"
)

// Application should describe your application.
type Application struct {
	name           string
	version        string
	gracefulPeriod time.Duration

	inventory  *scaffolder.Inventory
	components []scaffolder.Component
}

// Default assign the default variables for the application component.
func (a *Application) Default() {
	a.name = os.Args[0]
	a.version = "0.0.0"
	a.gracefulPeriod = time.Second
	a.inventory = scaffolder.New()
}

// String implements the Stringer interface.
func (a *Application) String() string {
	return fmt.Sprintf("%s (%s)", a.name, a.version)
}

// New build an application and customize it with the given Options.
func New(opts ...scaffolder.Option) (*Application, error) {
	app := &Application{}
	return app, scaffolder.Init(app, opts...)
}

// WithGracefulPeriod set the grace period allocated for stopping a component.
// The default value is one second.
func WithGracefulPeriod(duration time.Duration) scaffolder.Option {
	return func(a *Application) error {
		a.gracefulPeriod = duration
		return nil
	}
}

// WithComponent is used to attach register a component in the application life cycle.
func WithComponent(component scaffolder.Component, opts ...scaffolder.Option) scaffolder.Option {
	return func(a *Application) error {
		a.components = append(a.components, component)
		a.inventory.Add(component, opts...)
		return nil
	}
}

// WithVersion set the version of the application, the default value is "0.0.0".
func WithVersion(version string) scaffolder.Option {
	return func(a *Application) error {
		a.version = version
		return nil
	}
}

// WithName set the application name, the default value is the binary name.
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

// Run the application until it receives an interruption signal, or until its context
// has been canceled or expired, or until an error has been returned from the Start callback.
//
// The application would return an error if it was unable to Add a component, links the components,
// validate the components, start the components or stop the components.
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

	for {
		select {
		case <-ctx.Done():
			return
		case <-signalC:
			return
		case err = <-runtimeErr:
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Validator define the interface for components which should be validated.
type Validator interface {
	Validate() error
}

// StartHook define the interface for long running process which must run in parallel
// with the application.
type StartHook interface {
	Start(context.Context) error
}

// StopHook define the interface for components which require a graceful shutdown.
type StopHook interface {
	Stop(context.Context) error
}
