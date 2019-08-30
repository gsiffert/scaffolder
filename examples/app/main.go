package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Vorian-Atreides/scaffolder/application"
)

type A struct{}

func (a *A) Validate() error {
	fmt.Println("Validate A")
	return nil
}

func (a *A) Start(ctx context.Context) error {
	fmt.Println("Start A")
	return nil
}

func (a *A) Stop(ctx context.Context) error {
	fmt.Println("Stop A")
	return nil
}

type B struct {
	A *A
}

func (b *B) Validate() error {
	fmt.Println("Validate B")
	return nil
}

func (b *B) Start(ctx context.Context) error {
	fmt.Println("Start B")
	return nil
}

func (b *B) Stop(ctx context.Context) error {
	time.Sleep(time.Second)
	fmt.Println("Stop B")
	return nil
}

func main() {
	app, err := application.New(
		application.WithComponent(&A{}),
		application.WithComponent(&B{}),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
