package main

import (
	"context"
	"log"
	"time"

	"github.com/Vorian-Atreides/scaffolder/application"
	"github.com/Vorian-Atreides/scaffolder/component/logger"
)

type A struct {
	Logger logger.Logger
}

func (a *A) Start(ctx context.Context) error {
	a.Logger = a.Logger.With("namespace", "A")
	a.Logger.Infof("Starting")
	return nil
}

func (a *A) Stop(ctx context.Context) error {
	a.Logger.Infof("Stopping")
	return nil
}

type B struct {
	Logger logger.Logger
}

func (b *B) Start(ctx context.Context) error {
	b.Logger = b.Logger.With("namespace", "B")
	b.Logger.Infof("Starting")
	return nil
}

func (b *B) Stop(ctx context.Context) error {
	time.Sleep(time.Second)
	b.Logger.Infof("Stopping")
	return nil
}

type C struct {
	Logger logger.Logger
}

func (c *C) Start(ctx context.Context) error {
	c.Logger = c.Logger.With("namespace", "C")
	c.Logger = logger.New(logger.WithLevel(logger.Error), logger.WithLog(c.Logger))

	c.Logger.Infof("Starting")
	c.Logger.Errorf("Wait, I cannot start !")
	return nil
}

func (c *C) Stop(ctx context.Context) error {
	time.Sleep(time.Second)
	c.Logger.Infof("Stopping")
	return nil
}

func main() {
	app, err := application.New(
		application.WithComponent(logger.New()),
		application.WithComponent(&A{}),
		application.WithComponent(&B{}),
		application.WithComponent(&C{}),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
