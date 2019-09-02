package scaffolder

import (
	"errors"
	"reflect"
)

var (
	// ErrInvalidTarget is returned if the target given in the Init function is not a pointer.
	ErrInvalidTarget = errors.New("the target must be a pointer")
	// ErrInvalidOption is returned if the given Option does not respect the prototype:
	// func(pointer) error.
	ErrInvalidOption = errors.New("the option does not respect the mandatory prototype")

	errUnmatchingTargetType = errors.New("unmatching target type and option argument")
)

// Defaulter optional interface which can be used to attach default values
// to a component.
type Defaulter interface {
	Default()
}

// Option are generic functor used to configure a component,
// it is intended to be used to set one field at a time.
//
//   func(f *Form) error {
//   	  f.Age = age
// 	  return nil
//   }
//
// the function should always respect the prototype: func(pointer) error
type Option interface{}

func validate(o Option, target reflect.Type) error {
	oType := reflect.TypeOf(o)
	switch {
	case oType == nil:
		fallthrough
	case oType.Kind() != reflect.Func:
		fallthrough
	case oType.NumIn() != 1:
		fallthrough
	case oType.NumOut() != 1:
		fallthrough
	case oType.In(0).Kind() != reflect.Ptr:
		fallthrough
	case !oType.Out(0).Implements(errorInterface):
		return ErrInvalidOption
	case oType.In(0) != target:
		return errUnmatchingTargetType
	}
	return nil
}

// Init will take care of initializing the given Component by first calling
// the default method, if the component implements the Defaulter interface.
//
// Afterward, it will iterate through the list of options
// and apply them one after another.
//
//   type Form struct {
//   	Age       int
//   	FirstName string
//   }
//
//   func (f *Form) Default() {
//   	f.FirstName = "FirstName"
//   }
//
//   func WithAge(age int) scaffolder.Option {
//   	return func(f *Form) error {
//   		f.Age = age
//   		return nil
//   	}
//   }
func Init(target Component, opts ...Option) error {
	targetType := reflect.TypeOf(target)
	if targetType.Kind() != reflect.Ptr {
		return ErrInvalidOption
	}
	if defaulter, ok := target.(Defaulter); ok {
		defaulter.Default()
	}

	targetValue := reflect.ValueOf(target)
	args := []reflect.Value{targetValue}

	for _, opt := range opts {
		err := validate(opt, targetType)
		switch {
		case err == errUnmatchingTargetType:
			continue
		case err != nil:
			return err
		}

		returnedValues := reflect.ValueOf(opt).Call(args)
		if err, _ := returnedValues[0].Interface().(error); err != nil {
			return err
		}
	}
	return nil
}

// Configuration define a generic interface to turn configuration structure into
// usable component in the Scaffolder framework.
type Configuration interface {
	Options() []Option
}

// Configure apply the Options returned by the Configuration.
//
//   type Config struct {
//   	FirstName string `json:"first_name"`
//   	Age       int    `json:"age"`
//   }
//
//   func (c *Config) Options() []scaffolder.Option {
//   	return []scaffolder.Option{
//   		WithAge(c.Age),
//   	}
//   }
func Configure(target Component, cfg Configuration) error {
	return Init(target, cfg.Options()...)
}
