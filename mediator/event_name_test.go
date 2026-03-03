package mediator_test

import (
	"reflect"
	"testing"

	"github.com/oesand/octo/mediator"
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
			want: "github.com/oesand/octo/mediator_test.SimpleStruct",
		},
		{
			name: "*Struct",
			typ:  reflect.TypeFor[*SimpleStruct](),
			want: "*github.com/oesand/octo/mediator_test.SimpleStruct",
		},
		{
			name: "**Struct",
			typ:  reflect.TypeFor[**SimpleStruct](),
			want: "**github.com/oesand/octo/mediator_test.SimpleStruct",
		},
		{
			name: "GenericStruct[SimpleStruct]",
			typ:  reflect.TypeFor[GenericStruct[SimpleStruct]](),
			want: "github.com/oesand/octo/mediator_test.GenericStruct[github.com/oesand/octo/mediator_test.SimpleStruct]",
		},
		{
			name: "*GenericStruct[*SimpleStruct]",
			typ:  reflect.TypeFor[*GenericStruct[*SimpleStruct]](),
			want: "*github.com/oesand/octo/mediator_test.GenericStruct[*github.com/oesand/octo/mediator_test.SimpleStruct]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mediator.AbsoluteEventName(tt.typ); got != tt.want {
				t.Errorf("AbsoluteEventName() = %v, want %v", got, tt.want)
			}
		})
	}
}
