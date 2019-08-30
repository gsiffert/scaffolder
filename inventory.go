package scaffolder

import (
	"errors"
	"reflect"
)

const (
	tag = "scaffolder"
)

type component struct {
	value Component
	name  string
}

func (c *component) Default() {
	c.name = reflect.TypeOf(c.value).Elem().Name()
}

func WithName(name string) Option {
	return func(c *component) error {
		c.name = name
		return nil
	}
}

type Inventory struct {
	components map[string]*component
	fields     map[string][]reflect.Value

	addErr error
}

func New() *Inventory {
	return &Inventory{
		components: make(map[string]*component),
		fields:     make(map[string][]reflect.Value),
	}
}

func (i *Inventory) fieldByTag(field reflect.StructField) (string, bool) {
	kind := field.Type.Kind()
	tagName := field.Tag.Get(tag)
	switch {
	case len(tagName) == 0:
		return "", false
	case kind != reflect.Ptr && kind != reflect.Interface:
		return "", false
	}
	return tagName, true
}

func (i *Inventory) fieldByType(field reflect.StructField) (string, bool) {
	fieldType := field.Type
	name := fieldType.Name()
	switch {
	case fieldType.Kind() == reflect.Ptr:
		name = fieldType.Elem().Name()
	case fieldType.Kind() != reflect.Interface:
		return "", false
	}

	return name, true
}

func (i *Inventory) Add(c Component, opts ...Option) *Inventory {
	if i.addErr != nil {
		return i
	}

	cType := reflect.TypeOf(c)
	if cType.Kind() != reflect.Ptr {
		i.addErr = errors.New("")
		return i
	}
	cValue := reflect.ValueOf(c).Elem()
	cType = cType.Elem()

	component := &component{value: c}
	if err := Init(component, opts...); err != nil {
		i.addErr = err
		return i
	}

	i.components[component.name] = component
	// Store the addressable field values for later injection.
	for y := 0; y < cType.NumField(); y++ {
		field := cType.Field(y)
		value := cValue.Field(y)
		if !value.CanSet() {
			continue
		}

		name, ok := i.fieldByTag(field)
		if !ok {
			name, ok = i.fieldByType(field)
		}
		if ok {
			i.fields[name] = append(
				[]reflect.Value{value},
				i.fields[name]...,
			)
		}
	}
	return i
}

func (i *Inventory) Compile() error {
	if i.addErr != nil {
		return i.addErr
	}
	for key, fields := range i.fields {
		component, ok := i.components[key]
		if !ok {
			return errors.New("")
		}
		cValue := reflect.ValueOf(component.value)
		for _, field := range fields {
			if field.IsNil() {
				field.Set(cValue)
			}
		}
	}
	return nil
}
