package pm

// Set represents a simple generic hash set backed by a Go map.
type Set[K comparable] map[K]struct{}

func (m *Set[K]) create() {
	if *m == nil {
		*m = make(Set[K])
	}
}

// Add inserts one or more keys into the set.
// Duplicate keys are ignored.
func (m *Set[K]) Add(keys ...K) {
	m.create()
	for _, key := range keys {
		(*m)[key] = struct{}{}
	}
}

// Del removes one or more keys from the set.
// If the set is nil or keys do not exist, it does nothing.
func (m *Set[K]) Del(keys ...K) {
	if *m == nil {
		return
	}
	for _, key := range keys {
		delete(*m, key)
	}
}

// Len returns the current number of elements
func (m *Set[K]) Len() int {
	if *m == nil {
		return 0
	}
	return len(*m)
}

// Has checks whether a key exists in the set.
// Returns false if the set is nil or the key is missing.
func (m *Set[K]) Has(key K) bool {
	if *m == nil {
		return false
	}
	_, has := (*m)[key]
	return has
}

// CopyFrom copies all elements from another set into the current one.
// If the receiver set is nil, it will be initialized.
func (m *Set[K]) CopyFrom(other Set[K]) {
	if len(other) == 0 {
		return
	}
	m.create()
	for k := range other {
		(*m)[k] = struct{}{}
	}
}

// Values returns a slice containing all elements in the set.
// The order of elements is not guaranteed.
func (m *Set[K]) Values() []K {
	if *m == nil {
		return nil
	}
	keys := make([]K, 0, len(*m))
	for k := range *m {
		keys = append(keys, k)
	}
	return keys
}
