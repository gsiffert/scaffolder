package scaffolder

import (
	"reflect"
)

// Component let you define your application into small, independent, reusable element.
// Think of a component as a service or aggregate of structure and logic that you want to share
// with the rest of your code base or with other application.
//
// Not everything is intended to be a component, but anything which would benefits from
// the scaffolder dependency injection, configuration or application life cycle should be
// thought as a component.
type Component interface{}

// Container extends the Component interface with meta data.
type Container interface {
	Name() string
	Component() Component
}

type container struct {
	value Component
	name  string
	t     reflect.Type
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

// WithName attach a name to a component, it is useful to give a name to a component
// if you intend to manually configure the dependency injection with the structure tags.
func WithName(name string) Option {
	return func(c *container) error {
		c.name = name
		return nil
	}
}
