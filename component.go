package scaffolder

import (
	"reflect"
)

// Component is synonym in this project to what class is in OOP.
// A component is an individual unit, which can be configured and imported into
// other project while not perverting its usage.
type Component interface{}

// Container overlay of the Component, a container is identifiable by its name.
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
//
// The inventory algorithm is unable to reliably guess which component should be injected
// on which field if the component type or interface is not unique in the inventory.
// In this use-case, you must use tagged field and named component.
func WithName(name string) Option {
	return func(c *container) error {
		c.name = name
		return nil
	}
}
