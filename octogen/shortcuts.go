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

// ScanForMediatr marks a function as a Mediatr injector target for code generation.
//
// When used inside a declaration function, this marker instructs the generator
// to automatically discover and inject structs that implement
// `mediatr.NotificationHandler[T]` or `mediatr.RequestHandler[T]`,
// or to replace matching constructors following the `New{StructName}` naming convention.
//
// Like Inject, this function must not be called at runtime — it panics intentionally.
// It is used only by the code generator to identify Mediatr injection targets.
func ScanForMediatr() {
	panic("octo: cannot use function, only for scan")
}
