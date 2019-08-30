package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/Vorian-Atreides/scaffolder"
	"github.com/Vorian-Atreides/scaffolder/application"
	"github.com/Vorian-Atreides/scaffolder/component/healthcheck"
)

type A struct {
	s      healthcheck.Status
	status chan healthcheck.Status
}

func (a *A) Default() {
	a.status = make(chan healthcheck.Status)
}

func (a *A) Status() <-chan healthcheck.Status {
	return a.status
}

func (a *A) Start(ctx context.Context) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case a.status <- a.s:
			}
		}
	}()

	go func() {
		for {
			i := rand.Intn(4)
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(i) * time.Second):
			}
			a.s = healthcheck.Status(i)
		}
	}()
	return nil
}

type B struct {
	HC *healthcheck.HealthChecker
}

func (b *B) Start(ctx context.Context) error {
	go func() {
		for {
			var report map[string]healthcheck.Status
			select {
			case <-ctx.Done():
				return
			case report = <-b.HC.Status():
			}
			fmt.Println(report)
		}
	}()
	return nil
}

func main() {
	app, err := application.New(
		application.WithComponent(&healthcheck.HealthChecker{}),
		application.WithComponent(&A{}),
		application.WithComponent(&A{}, scaffolder.WithName("a2")),
		application.WithComponent(&B{}),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
