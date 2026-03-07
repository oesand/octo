package octogen

// Inject marks a struct type or constructor function for dependency injection code generation.
//
// Usage examples:
//
//	octogen.Inject[*MyStruct]()                 // marks a struct for auto-construction
//	octogen.Inject[*MyStruct]("key1")           // marks a named injection target
//	octogen.Inject(NewMyStruct)                 // marks a constructor function for injection
//	octogen.Inject(NewMyStruct, "key2")         // marks a named constructor injection
//
// During code generation, all Inject calls inside declaration functions
// are replaced with concrete `octo.Inject` or `octo.InjectNamed` invocations
// that register constructors in the container.
//
// This function must never be called at runtime — it panics intentionally.
// It is used only as a compile-time marker for the generator.
func Inject[T any](...any) {
	panic("octo: cannot use function, only for scan")
}

// Fields marks a struct type for fields code generation.
func Fields[T any]() {
	panic("octo: cannot use function, only for scan")
}
