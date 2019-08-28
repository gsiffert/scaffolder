package main

import (
	"fmt"

	"github.com/Vorian-Atreides/scaffolder"
)

type A struct {
	C *C `scaffolder:"c"`
}

type B struct {
	A *A `scaffolder:"a"`
}

type C struct {
	A *A          `scaffolder:"a"`
	B interface{} `scaffolder:"b"`
}

func main() {
	inventory := scaffolder.New()
	a := A{}
	b := B{}
	c := C{}

	inventory.Add(&a, "a").Add(&b, "b").Add(&c, "c")
	err := inventory.Compile()
	fmt.Printf("A: %v\nB: %v\nC: %v\nErr: %v\n", a, b, c, err)
}
