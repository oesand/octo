package mdtreflect

import (
	"fmt"
	"reflect"
	"testing"
)

type SimpleStruct struct{}

func TestAbsoluteTypeName(t *testing.T) {
	tests := []struct {
		name string
		typ  reflect.Type
		want string
	}{
		{
			name: "Builtin type",
			typ:  reflect.TypeFor[string](),
			want: "string",
		},
		{
			name: "Fmt package",
			typ:  reflect.TypeFor[fmt.Stringer](),
			want: "fmt/Stringer",
		},
		{
			name: "Struct",
			typ:  reflect.TypeFor[SimpleStruct](),
			want: "github.com/oesand/octo/mediatr/mdtreflect/SimpleStruct",
		},
		{
			name: "*Struct",
			typ:  reflect.TypeFor[*SimpleStruct](),
			want: "*github.com/oesand/octo/mediatr/mdtreflect/SimpleStruct",
		},
		{
			name: "**Struct",
			typ:  reflect.TypeFor[**SimpleStruct](),
			want: "**github.com/oesand/octo/mediatr/mdtreflect/SimpleStruct",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AbsoluteTypeName(tt.typ); got != tt.want {
				t.Errorf("AbsoluteTypeName() = %v, want %v", got, tt.want)
			}
		})
	}
}
