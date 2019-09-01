package scaffolder

import (
	"errors"
	"reflect"
)

const (
	tag = "scaffolder"
	all = "containers"
)

var (
	// ErrInvalidComponent means that the given component was neither a pointer not an interface.
	ErrInvalidComponent = errors.New("component is neither a pointer nor an interface")
)

type field struct {
	value reflect.Value
	t     reflect.Type
	tag   string
	name  string
}

// Inventory does the magic of wiring your components together.
// Any component which has been Added to the Inventory becomes a candidate to be injected
// in the other components hold in the Inventory.
//
// Currently the assignment algorithm follow this priority:
// 1. The component name match the field tag.
// 2. The component name and type match the field name and type (case sensitive).
// 3. The component type match the field type.
// 4. Component implements the field interface.
//
// A valid assignable field must be a public and can be either a pointer, an interface or a slice.
// The slice is currently an exception and only works for the tag `scaffolder:"all"`, which will
// inject an array with every components hold in the inventory.
type Inventory struct {
	fields     []field
	containers []*container
	all        []Container

	addErr error
}

// New build a new Inventory.
func New() *Inventory {
	return &Inventory{}
}

func (i *Inventory) isSettableType(kind reflect.Kind) bool {
	return kind == reflect.Slice || kind == reflect.Ptr || kind == reflect.Interface
}

func (i *Inventory) extractFields(container *container) []field {
	componentType := container.t.Elem()
	if componentType.Kind() != reflect.Struct {
		return nil
	}

	var fields []field
	structType := componentType
	structValue := reflect.ValueOf(container.Component()).Elem()
	for y := 0; y < structType.NumField(); y++ {
		fieldType := structType.Field(y)
		fieldValue := structValue.Field(y)
		if !fieldValue.CanSet() || !i.isSettableType(fieldValue.Type().Kind()) {
			continue
		}

		f := field{
			value: fieldValue,
			t:     fieldValue.Type(),
			tag:   fieldType.Tag.Get(tag),
			name:  fieldType.Name,
		}
		fields = append(fields, f)
	}
	return fields
}

// Add a component to the inventory, it will take care of calling Init with the
// given options. You should use the WithName Option if you intend to assign
// this component to matching structure tags.
func (i *Inventory) Add(component Component, opts ...Option) *Inventory {
	if i.addErr != nil {
		return i
	}

	cType := reflect.TypeOf(component)
	if cType.Kind() != reflect.Ptr && cType.Kind() != reflect.Interface {
		i.addErr = ErrInvalidComponent
		return i
	}
	if err := Init(component, opts...); err != nil {
		i.addErr = err
		return i
	}

	container := &container{value: component, t: cType}
	if err := Init(container, opts...); err != nil {
		i.addErr = err
		return i
	}

	i.containers = append(i.containers, container)
	i.all = append(i.all, container)
	i.fields = append(i.fields, i.extractFields(container)...)
	return i
}

// We prioritize the assignement by:
// 1. The component name match the field tag
// 2. The component name and type match the field name and type.
// 3. The component type match the field type.
// 4. Component implements the field interface.
var conditions = []func(field field, container *container) bool{
	func(field field, container *container) bool {
		return field.tag == container.name
	},
	func(field field, container *container) bool {
		return field.t == container.t && field.name == container.name
	},
	func(field field, container *container) bool {
		return field.t == container.t
	},
	func(field field, container *container) bool {
		return field.t.Kind() == reflect.Interface &&
			container.t.Implements(field.t)
	},
}

// Compile will perform the linking for every components which have been added in the inventory.
func (i *Inventory) Compile() error {
	if i.addErr != nil {
		return i.addErr
	}

	// This algorithm has a time complexity of O(N)3,
	// it assume that no inventory will ever have hundreds of components.
Next:
	for _, field := range i.fields {
		if !field.value.IsNil() {
			continue
		}

		for _, condition := range conditions {
			for _, container := range i.containers {
				if condition(field, container) {
					value := reflect.ValueOf(container.value)
					field.value.Set(value)
					continue Next
				}
			}
		}
	}
	return nil
}
