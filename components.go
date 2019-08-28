package scaffolder

import (
	"errors"
	"reflect"
)

type Component interface{}

const (
	tag = "scaffolder"
)

type Inventory struct {
	components map[string]Component
	fields     map[string][]reflect.Value

	addErr error
}

func New() *Inventory {
	return &Inventory{
		components: make(map[string]Component),
		fields:     make(map[string][]reflect.Value),
	}
}

func (i *Inventory) Add(component Component, name string) *Inventory {
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

	i.components[name] = component
	// Store the addressable field values for later injection.
	for y := 0; y < cType.NumField(); y++ {
		field := cType.Field(y)
		tagName := field.Tag.Get(tag)
		if len(tagName) > 0 {
			switch field.Type.Kind() {
			case reflect.Ptr:
				break
			case reflect.Interface:
				break
			default:
				continue
			}
			fieldValue := cValue.Field(y)
			if fieldValue.CanSet() {
				i.fields[tagName] = append(
					[]reflect.Value{fieldValue},
					i.fields[tagName]...,
				)
			}
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
		cValue := reflect.ValueOf(component)
		for _, field := range fields {
			field.Set(cValue)
		}
	}
	return nil
}
