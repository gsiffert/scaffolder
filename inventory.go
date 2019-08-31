package scaffolder

import (
	"errors"
	"reflect"
)

import "fmt"

const (
	tag = "scaffolder"
	all = "containers"
)

type Inventory struct {
	containerByTags  map[string]*container
	containerByTypes map[reflect.Type]*container
	fieldByTags      map[string][]reflect.Value
	fieldByTypes     map[reflect.Type][]reflect.Value

	addErr error
	all    []Container
}

func New() *Inventory {
	return &Inventory{
		containerByTags:  make(map[string]*container),
		containerByTypes: make(map[reflect.Type]*container),
		fieldByTags:      make(map[string][]reflect.Value),
		fieldByTypes:     make(map[reflect.Type][]reflect.Value),
	}
}

func (i *Inventory) fieldByTag(field reflect.StructField) (string, bool) {
	tagName := field.Tag.Get(tag)
	return tagName, len(tagName) > 0
}

func (i *Inventory) isSettableType(kind reflect.Kind) bool {
	return kind == reflect.Slice || kind == reflect.Ptr || kind == reflect.Interface
}

func (i *Inventory) Add(component Component, opts ...Option) *Inventory {
	if i.addErr != nil {
		return i
	}

	cType := reflect.TypeOf(component)
	if cType.Kind() != reflect.Ptr && cType.Kind() != reflect.Interface {
		i.addErr = errors.New("B")
		return i
	}

	container := &container{value: component}
	if err := Init(container, opts...); err != nil {
		i.addErr = err
		return i
	}

	i.all = append(i.all, container)
	i.containerByTags[container.name] = container
	i.containerByTypes[cType] = container

	cValue := reflect.ValueOf(component).Elem()
	cType = cType.Elem()
	// Store the addressable field values for later injection.
	for y := 0; y < cType.NumField(); y++ {
		field := cType.Field(y)
		value := cValue.Field(y)
		if !value.CanSet() && !i.isSettableType(field.Type.Kind()) {
			continue
		}

		if name, ok := i.fieldByTag(field); ok {
			i.fieldByTags[name] = append(
				[]reflect.Value{value},
				i.fieldByTags[name]...,
			)
		}
		i.fieldByTypes[field.Type] = append(
			[]reflect.Value{value},
			i.fieldByTypes[field.Type]...,
		)
	}
	return i
}

func (i *Inventory) Compile() error {
	if i.addErr != nil {
		return i.addErr
	}
	i.containerByTags[all] = &container{value: i.all}

	// Attempt to assign by tag, any missing tag is considered as a fatal error
	// because it was explicitly requested to be found by the user.
	for key, fields := range i.fieldByTags {
		container, ok := i.containerByTags[key]
		if !ok {
			return errors.New("tag not found")
		}
		cValue := reflect.ValueOf(container.value)
		for _, field := range fields {
			if field.IsNil() {
				field.Set(cValue)
			}
		}
	}

	// Attempt to assign by struct Type and fallback to interface implementation
	// if no type is found.
	for key, fields := range i.fieldByTypes {
		container, ok := i.containerByTypes[key]
		if !ok {
			for t, c := range i.containerByTypes {
				if key.Kind() == reflect.Interface && t.Implements(key) {
					fmt.Println("A")
					container = c
					ok = true
					break
				}
			}
			if !ok {
				continue
			}
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
