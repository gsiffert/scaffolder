package scaffolder_test

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

type SomeComponent struct {
	HealthRegistry healthcheck.HealthRegistry
}

func (s *SomeComponent) Start(ctx context.Context) error {
	for {
		i := rand.Intn(4)
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Duration(i) * time.Second):
		}
		a.HealthRegistry.SetStatus("SomeComponent", healthcheck.Status(i))
	}
	return nil
}

type Server struct {
	Readiness *healthcheck.HTTPHandler `scaffolder:"readiness"`
	Liveness  *healthcheck.HTTPHandler `scaffolder:"liveness"`
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.DefaultServeMux
	mux.Handle("/ready", s.Readiness)
	mux.Handle("/healthy", s.Liveness)
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	server.ListenAndServe()
	return nil
}

func Example() {
	app, err := application.New(
		application.WithComponent(healthcheck.NewRegistry()),
		application.WithComponent(&SomeComponent{}),
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
		application.WithComponent(&Server{}),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
