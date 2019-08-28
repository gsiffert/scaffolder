package scaffolder

import (
	"errors"
	"reflect"
)

type Defaulter interface {
	Default()
}

type Option interface{}

func validate(o Option) error {
	oType := reflect.TypeOf(o)
	switch {
	case oType == nil:
		return errors.New("f")
	case oType.Kind() != reflect.Func:
		return errors.New("e")
	case oType.NumIn() != 1:
		return errors.New("d")
	case oType.NumOut() != 1:
		return errors.New("c")
	case oType.In(0).Kind() != reflect.Ptr:
		return errors.New("b")
	case !oType.Out(0).Implements(errorInterface):
		return errors.New("a")
	}
	return nil
}

func Options(target interface{}, opts ...Option) error {
	targetType := reflect.TypeOf(target)
	if targetType.Kind() != reflect.Ptr {
		return errors.New("")
	}
	if defaulter, ok := target.(Defaulter); ok {
		defaulter.Default()
	}

	targetValue := reflect.ValueOf(target)
	args := []reflect.Value{targetValue}

	for _, opt := range opts {
		if err := validate(opt); err != nil {
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
	return Options(target, cfg.Options()...)
}
