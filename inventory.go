package scaffolder

import (
	"errors"
	"reflect"
)

const (
	tag = "scaffolder"
	all = "containers"
)

type Inventory struct {
	containers map[string]*container
	fields     map[string][]reflect.Value

	addErr error
	all    []Container
}

func New() *Inventory {
	return &Inventory{
		containers: make(map[string]*container),
		fields:     make(map[string][]reflect.Value),
	}
}

func (i *Inventory) fieldByTag(field reflect.StructField) (string, bool) {
	kind := field.Type.Kind()
	tagName := field.Tag.Get(tag)
	switch {
	case len(tagName) == 0:
		return "", false
	case kind != reflect.Ptr && kind != reflect.Interface && kind != reflect.Slice:
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
	case fieldType.Kind() != reflect.Interface && fieldType.Kind() != reflect.Slice:
		return "", false
	}

	return name, true
}

func (i *Inventory) Add(component Component, opts ...Option) *Inventory {
	if i.addErr != nil {
		return i
	}

	cType := reflect.TypeOf(component)
	if cType.Kind() != reflect.Ptr {
		i.addErr = errors.New("")
		return i
	}
	cValue := reflect.ValueOf(component).Elem()
	cType = cType.Elem()

	container := &container{value: component}
	if err := Init(container, opts...); err != nil {
		i.addErr = err
		return i
	}

	i.all = append(i.all, container)
	i.containers[container.name] = container
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
	i.containers[all] = &container{value: i.all}

	for key, fields := range i.fields {
		container, ok := i.containers[key]
		if !ok {
			return errors.New("")
		}
		cValue := reflect.ValueOf(container.value)
		for _, field := range fields {
			if field.IsNil() {
				field.Set(cValue)
			}
		}
	}
	return nil
}
