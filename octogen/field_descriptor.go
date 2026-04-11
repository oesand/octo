package octogen

type FieldDescriptor[Struct any, Field any] struct {
	Name  string
	Value func(*Struct) Field
}

func (desc FieldDescriptor[Struct, Field]) GetName() string {
	return desc.Name
}

func (desc FieldDescriptor[Struct, Field]) GetValue(s *Struct) Field {
	return desc.Value(s)
}
