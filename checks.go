package octo

import (
	"reflect"
)

var containerPtrType = reflect.TypeFor[*Container]()

func ensureCanInjectType(typ reflect.Type) {
	if typ.AssignableTo(containerPtrType) {
		panic("cannot inject Container")
	}
}
