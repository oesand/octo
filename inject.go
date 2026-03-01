package octo

func ensureCanInjectType[T any]() {
	var t T
	switch any(t).(type) {
	case *Container:
		panic("cannot inject Container")
	}
}

// TryInjectValue registers a concrete value into the container if not registered.
func TryInjectValue[T any](container *Container, value T) bool {
	return TryInjectNamedValue[T](container, "", value)
}

// TryInjectNamedValue registers a concrete value with a name into the container if not registered.
func TryInjectNamedValue[T any](container *Container, name string, value T) bool {
	ensureCanInjectType[T]()

	container = containerOrDefault(container)
	container.mu.Lock()
	defer container.mu.Unlock()

	if resolve[T](container, name) != nil {
		return false
	}

	injectValue(container, name, value)
	return true
}

// InjectValue registers a concrete value into the container.
func InjectValue[T any](container *Container, value T) {
	InjectNamedValue[T](container, "", value)
}

// InjectNamedValue registers a concrete value with a name for named resolution.
func InjectNamedValue[T any](container *Container, name string, value T) {
	ensureCanInjectType[T]()

	container = containerOrDefault(container)
	container.mu.Lock()
	defer container.mu.Unlock()

	injectValue(container, name, value)
}

// TryInject registers a provider function to lazily resolve a type if not registered.
func TryInject[T any](container *Container, provider Provider[T]) bool {
	return TryInjectNamed(container, "", provider)
}

// TryInjectNamed registers a named provider function to lazily resolve a type if not registered.
func TryInjectNamed[T any](container *Container, name string, provider Provider[T]) bool {
	ensureCanInjectType[T]()

	container = containerOrDefault(container)
	container.mu.Lock()
	defer container.mu.Unlock()

	if resolve[T](container, name) != nil {
		return false
	}

	injectLazy(container, name, provider)
	return true
}

// Inject registers a provider function to lazily resolve a type.
func Inject[T any](container *Container, provider Provider[T]) {
	InjectNamed[T](container, "", provider)
}

// InjectNamed registers a named provider function to lazily resolve a type.
func InjectNamed[T any](container *Container, name string, provider Provider[T]) {
	ensureCanInjectType[T]()

	container = containerOrDefault(container)
	container.mu.Lock()
	defer container.mu.Unlock()

	injectLazy(container, name, provider)
}
