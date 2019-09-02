package scaffolder

import (
	"reflect"
)

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()
