package mediatr

import (
	"reflect"
)

// AbsoluteEventName returns the "absolute" name of a type including:
// 1. The package import path (PkgPath)
// 2. The struct name
// 3. The pointer levels (e.g., "*", "**", etc.)
//
// Examples:
//   - type MyStruct struct{} in package "github.com/user/project/pkg"
//   - AbsoluteEventName(MyStruct)       => "github.com/user/project/pkg/MyStruct"
//   - AbsoluteEventName(*MyStruct)      => "*github.com/user/project/pkg/MyStruct"
//   - AbsoluteEventName(**MyStruct)     => "**github.com/user/project/pkg/MyStruct"
func AbsoluteEventName(typ reflect.Type) string {
	target := typ
	var ptrLevel string
	for target.Kind() == reflect.Pointer {
		target = target.Elem()
		ptrLevel += "*"
	}

	if target.Kind() != reflect.Struct {
		panic("unsupported type for event: " + target.String())
	}

	name := target.Name()
	if pkg := target.PkgPath(); pkg != "" {
		name = pkg + "." + name
	}
	if ptrLevel != "" {
		name = ptrLevel + name
	}

	return name
}
