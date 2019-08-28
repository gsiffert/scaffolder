package application

import (
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Vorian-Atreides/scaffolder"
)

type Application struct {
	name    string
	version string

	inventory  *scaffolder.Inventory
	components map[string]scaffolder.Component
}

func (a *Application) Default() {
	a.name = os.Args[0]
	a.version = "0.0.0"
	a.inventory = scaffolder.New()
	a.components = make(map[string]scaffolder.Component)
}

func WithComponent(component scaffolder.Component, name string, opts ...scaffolder.Option) scaffolder.Option {
	return func(a *Application) error {
		if err := scaffolder.Options(component, opts); err != nil {
			return err
		}
		if _, ok := a.components[name]; ok {
			return errors.New("")
		}
		if err := scaffolder.Options(component, opts...); err != nil {
			return err
		}
		a.components[name] = component
		a.inventory.Add(component, name)
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

func (a *Application) Run() error {
	if err := a.inventory.Compile(); err != nil {
		return err
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	wg := sync.WaitGroup{}
	defer wg.Wait()

	for _, component := range a.components {
		// Validate the components before starting them.
		if validator, ok := component.(Validator); ok {
			if err := validator.Validate(); err != nil {
				return err
			}
		}
		// Start the components.
		if s, ok := component.(StartHook); ok {
			if err := s.Start(); err != nil {
				return err
			}
			wg.Add(1)
		}
		// Hook the call to stop the component when shutting down the application.
		if s, ok := component.(StopHook); ok {
			defer func() {
				_ = s.Stop()
				wg.Done()
			}()
		}
	}

	<-c
}

type Validator interface {
	Validate() error
}

type StartHook interface {
	Start() error
}

type StopHook interface {
	Stop() error
}
