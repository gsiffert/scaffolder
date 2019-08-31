package scaffolder

import (
	"errors"
	"reflect"
)

var (
	ErrInvalidTarget        = errors.New("the target must be a pointer")
	ErrInvalidOption        = errors.New("the option does not respect the mandatory prototype")
	errUnmatchingTargetType = errors.New("unmatching target type and option argument")
)

type Defaulter interface {
	Default()
}

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

type Configuration interface {
	Options() []Option
}

func Configure(target interface{}, cfg Configuration) error {
	return Init(target, cfg.Options()...)
}
