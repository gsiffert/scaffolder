package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"github.com/Vorian-Atreides/scaffolder"
	"github.com/Vorian-Atreides/scaffolder/application"
	"github.com/Vorian-Atreides/scaffolder/component/healthcheck"
)

type A struct {
	defaultHC healthcheck.DefaultHealthCheck
}

func (a *A) Default() {
	a.defaultHC = healthcheck.DefaultClient(context.Background())
}

func (a *A) Status() <-chan healthcheck.Status {
	return a.defaultHC.Status()
}

func (a *A) Start(ctx context.Context) error {
	fmt.Println("A")
	go func() {
		for {
			rand.Intn(4)
			select {
			case <-ctx.Done():
				return
			default:
				//case <-time.After(time.Duration(i) * time.Second):
			}
			a.defaultHC.SetStatus(healthcheck.Ready)
			// a.defaultHC.SetStatus(healthcheck.Status(i))
		}
	}()
	return nil
}

type B struct {
	Readiness *healthcheck.HTTPHandler
	//Liveness  *healthcheck.HTTPHandler
}

func (b *B) Start(ctx context.Context) error {
	mux := http.DefaultServeMux
	mux.Handle("/ready", b.Readiness)
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	fmt.Println("B")
	go func() {
		server.ListenAndServe()
	}()
	return nil
}

func main() {
	app, err := application.New(
		application.WithComponent(&healthcheck.HealthChecker{}),
		application.WithComponent(&A{}),
		application.WithComponent(&A{}, scaffolder.WithName("a2")),
		application.WithComponent(&healthcheck.HTTPHandler{}),
		application.WithComponent(&B{}),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
