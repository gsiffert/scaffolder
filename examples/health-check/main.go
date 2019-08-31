package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/Vorian-Atreides/scaffolder"
	"github.com/Vorian-Atreides/scaffolder/application"
	"github.com/Vorian-Atreides/scaffolder/component/healthcheck"
)

type A struct {
	HealthRegistry healthcheck.HealthRegistry
	name           string
}

func (a *A) Start(ctx context.Context) error {
	for {
		i := rand.Intn(4)
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Duration(i) * time.Second):
		}
		a.HealthRegistry.SetStatus(a.name, healthcheck.Status(i))
	}
	return nil
}

type B struct {
	Readiness *healthcheck.HTTPHandler `scaffolder:"readiness"`
	Liveness  *healthcheck.HTTPHandler `scaffolder:"liveness"`
}

func (b *B) Start(ctx context.Context) error {
	mux := http.DefaultServeMux
	mux.Handle("/ready", b.Readiness)
	mux.Handle("/healthy", b.Liveness)
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	server.ListenAndServe()
	return nil
}

func main() {
	app, err := application.New(
		application.WithComponent(healthcheck.NewRegistry()),
		application.WithComponent(&A{name: "a"}),
		application.WithComponent(&A{name: "a2"}, scaffolder.WithName("a2")),
		application.WithComponent(
			&healthcheck.HTTPHandler{},
			scaffolder.WithName("readiness"),
		),
		application.WithComponent(
			&healthcheck.HTTPHandler{},
			healthcheck.WithMergingStrategy(
				healthcheck.EveryService(healthcheck.Healthy, healthcheck.NotReady),
			),
			scaffolder.WithName("liveness"),
		),
		application.WithComponent(&B{}),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
