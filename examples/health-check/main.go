package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	//"github.com/Vorian-Atreides/scaffolder"
	"github.com/Vorian-Atreides/scaffolder/application"
	"github.com/Vorian-Atreides/scaffolder/component/healthcheck"
)

type A struct {
	//HealthRegistry healthcheck.HealthRegistry
}

func (a *A) Start(ctx context.Context) error {
	go func() {
		for {
			i := rand.Intn(4)
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(i) * time.Second):
			}
			//a.HealthRegistry.SetStatus("a", healthcheck.Ready)
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
		application.WithComponent(healthcheck.NewRegistry()),
		application.WithComponent(&A{}),
		//application.WithComponent(&A{}, scaffolder.WithName("a2")),
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
