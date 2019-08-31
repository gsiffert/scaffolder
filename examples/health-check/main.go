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
	defaultHC healthcheck.DefaultHealthCheck
}

func (a *A) Default() {}

func (a *A) Status() <-chan healthcheck.Status {
	return a.defaultHC.Status()
}

func (a *A) Start(ctx context.Context) error {
	a.defaultHC = healthcheck.DefaultClient(ctx)

	go func() {
		for {
			i := rand.Intn(4)
			select {
			case <-ctx.Done():
				return
			default:
				//case <-time.After(time.Duration(i) * time.Second):
			}
			a.defaultHC.SetStatus(healthcheck.Status(i))
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
