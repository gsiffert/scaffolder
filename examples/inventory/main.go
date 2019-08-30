package main

import (
	"fmt"

	"github.com/Vorian-Atreides/scaffolder"
)

type A struct {
}

type B struct {
	A *A `scaffolder:"a"`
}

type C struct {
	A *A `scaffolder:"a"`
	B *B
}

func main() {
	inventory := scaffolder.New()
	a := A{}
	b := B{}
	c := C{}

	inventory.
		Add(&a, scaffolder.WithName("a")).
		Add(&b).
		Add(&c)
	err := inventory.Compile()
	fmt.Printf("A: %v\nB: %v\nC: %v\nErr: %v\n", a, b, c, err)
}
