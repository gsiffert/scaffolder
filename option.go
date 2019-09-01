// package scaffolder define the smallest components from the Scaffolder framework.
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
	ErrInvalidOption        = errors.New("the option does not respect the mandatory prototype")
	errUnmatchingTargetType = errors.New("unmatching target type and option argument")
)

// Defaulter optional interface which can be implemented to attach default values
// to a component.
type Defaulter interface {
	Default()
}

// Option generic functor used to configure a component.
// It should be used to be side-effect free and focus on setting only one field at a time.
// The only supported prototype is: func(pointer) error
// where pointer must be a pointer to your component.
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

// Init will first try to call the Default function if the target implements
// the Defaulter interface. Afterward, it will iterate through the list of options
// and apply them one after another.
func Init(target interface{}, opts ...Option) error {
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
func Configure(target interface{}, cfg Configuration) error {
	return Init(target, cfg.Options()...)
}
