package mediatr

import (
	"reflect"
	"testing"
)

type SimpleStruct struct{}
type GenericStruct[T any] struct{}

func TestAbsoluteTypeName(t *testing.T) {
	tests := []struct {
		name string
		typ  reflect.Type
		want string
	}{
		{
			name: "Struct",
			typ:  reflect.TypeFor[SimpleStruct](),
			want: "github.com/oesand/octo/mediatr.SimpleStruct",
		},
		{
			name: "*Struct",
			typ:  reflect.TypeFor[*SimpleStruct](),
			want: "*github.com/oesand/octo/mediatr.SimpleStruct",
		},
		{
			name: "**Struct",
			typ:  reflect.TypeFor[**SimpleStruct](),
			want: "**github.com/oesand/octo/mediatr.SimpleStruct",
		},
		{
			name: "GenericStruct[SimpleStruct]",
			typ:  reflect.TypeFor[GenericStruct[SimpleStruct]](),
			want: "github.com/oesand/octo/mediatr.GenericStruct[github.com/oesand/octo/mediatr.SimpleStruct]",
		},
		{
			name: "*GenericStruct[*SimpleStruct]",
			typ:  reflect.TypeFor[*GenericStruct[*SimpleStruct]](),
			want: "*github.com/oesand/octo/mediatr.GenericStruct[*github.com/oesand/octo/mediatr.SimpleStruct]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AbsoluteEventName(tt.typ); got != tt.want {
				t.Errorf("AbsoluteEventName() = %v, want %v", got, tt.want)
			}
		})
	}
}
