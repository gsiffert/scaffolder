package scaffolder

import (
	"reflect"
)

type Component interface {
}

type Container interface {
	Name() string
	Component() Component
}

type container struct {
	value Component
	name  string
}

func (c *container) Default() {
	c.name = reflect.TypeOf(c.value).Elem().Name()
}

func (c *container) Name() string {
	return c.name
}

func (c *container) Component() Component {
	return c.value
}

func WithName(name string) Option {
	return func(c *container) error {
		c.name = name
		return nil
	}
}
