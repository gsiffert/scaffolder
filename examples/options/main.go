package main

import (
	"encoding/json"
	"fmt"

	"github.com/Vorian-Atreides/scaffolder"
)

type Form struct {
	FirstName string
	LastName  string
	Age       int
}

func (f *Form) Default() {
	f.FirstName = "FirstName"
	f.LastName = "LastName"
	f.Age = 42
}

func FirstName(value string) scaffolder.Option {
	return func(f *Form) error {
		f.FirstName = value
		return nil
	}
}

func LastName(value string) scaffolder.Option {
	return func(f *Form) error {
		f.LastName = value
		return nil
	}
}

func Age(value int) scaffolder.Option {
	return func(f *Form) error {
		f.Age = value
		return nil
	}
}

const data = `{"first_name": "Erika", "last_name": "Matsukawa", "age": 28}`

type Config struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Age       int    `json:"age"`
}

func (c *Config) Options() []scaffolder.Option {
	return []scaffolder.Option{
		FirstName(c.FirstName),
		LastName(c.LastName),
		Age(c.Age),
	}
}

func main() {
	var form Form
	_ = scaffolder.Options(&form)
	fmt.Printf("Form: %v\n", form)

	_ = scaffolder.Options(
		&form,
		FirstName("Gaston"),
		LastName("Siffert"),
		Age(27),
	)
	fmt.Printf("Form: %v\n", form)

	var cfg Config
	_ = json.Unmarshal([]byte(data), &cfg)
	_ = scaffolder.Configure(&form, &cfg)
	fmt.Printf("Form: %v\n", form)
}
