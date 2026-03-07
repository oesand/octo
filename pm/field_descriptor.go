package pm

type Field[T any] interface {
	GetName() string
	GetValue(*T) any
}

type FieldDescriptor[T any, F any] struct {
	Name  string
	Value func(*T) F
}

func (desc FieldDescriptor[T, F]) GetName() string {
	return desc.Name
}

func (desc FieldDescriptor[T, F]) GetValue(s *T) any {
	return desc.Value(s)
}
