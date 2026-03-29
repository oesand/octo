package internal

import "testing"

type Struct struct {
}

func (s *Struct) Name() string { return "" }

type Iface interface {
	Name() string
}

type EmptyIface interface {
}

func TestType_Real(t1 *testing.T) {
	type testCase struct {
		Name string
		Type ShadowType
		Want bool
	}
	tests := []testCase{
		{
			Name: "interface",
			Type: Type[Iface]{},
			Want: false,
		},
		{
			Name: "empty interface",
			Type: Type[EmptyIface]{},
			Want: false,
		},
		{
			Name: "slice interface",
			Type: Type[[]Iface]{},
			Want: true,
		},
		{
			Name: "pointer interface",
			Type: Type[*Iface]{},
			Want: true,
		},
		{
			Name: "struct",
			Type: Type[Struct]{},
			Want: true,
		},
		{
			Name: "pointer struct",
			Type: Type[*Struct]{},
			Want: true,
		},
		{
			Name: "slice struct",
			Type: Type[[]Struct]{},
			Want: true,
		},
		{
			Name: "slice of pointer struct",
			Type: Type[[]*Struct]{},
			Want: true,
		},
		{
			Name: "int",
			Type: Type[int]{},
			Want: true,
		},
		{
			Name: "float32",
			Type: Type[float32]{},
			Want: true,
		},
		{
			Name: "string",
			Type: Type[string]{},
			Want: true,
		},
		{
			Name: "chan",
			Type: Type[chan int]{},
			Want: true,
		},
		{
			Name: "map",
			Type: Type[map[Iface]EmptyIface]{},
			Want: true,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.Name, func(t1 *testing.T) {
			if got := tt.Type.Real(); got != tt.Want {
				t1.Errorf("Real() = %v, want %v", got, tt.Want)
			}
		})
	}
}

func TestType_ConvertibleFrom(t1 *testing.T) {
	type testCase struct {
		name string
		from ShadowType
		to   ShadowType
		want bool
	}
	tests := []testCase{
		{
			name: "same interface",
			from: Type[Iface]{},
			to:   Type[Iface]{},
			want: true,
		},
		{
			name: "to interface, struct not implement",
			from: Type[Struct]{},
			to:   Type[Iface]{},
			want: false,
		},
		{
			name: "to interface, pointer struct implement",
			from: Type[*Struct]{},
			to:   Type[Iface]{},
			want: true,
		},
		{
			name: "to interface, string",
			from: Type[string]{},
			to:   Type[Iface]{},
			want: false,
		},
		{
			name: "to interface, int",
			from: Type[int]{},
			to:   Type[Iface]{},
			want: false,
		},
		{
			name: "empty interface, struct",
			from: Type[Struct]{},
			to:   Type[EmptyIface]{},
			want: true,
		},
		{
			name: "empty interface, pointer struct",
			from: Type[*Struct]{},
			to:   Type[EmptyIface]{},
			want: true,
		},
		{
			name: "empty interface, string",
			from: Type[string]{},
			to:   Type[EmptyIface]{},
			want: true,
		},
		{
			name: "empty interface, int",
			from: Type[int]{},
			to:   Type[EmptyIface]{},
			want: true,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			if got := tt.to.ConvertibleFrom(tt.from); got != tt.want {
				t1.Errorf("ConvertibleFrom() = %v, want %v", got, tt.want)
			}
		})
	}
}
