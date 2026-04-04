package octogen

type Field[Struct any] interface {
	GetName() string
	GetValue(*Struct) any
}

type FieldDescriptor[Struct any, Field any] struct {
	Name  string
	Value func(*Struct) Field
}

func (desc FieldDescriptor[Struct, F]) GetName() string {
	return desc.Name
}

func (desc FieldDescriptor[Struct, F]) GetValue(s *Struct) any {
	return desc.Value(s)
}
