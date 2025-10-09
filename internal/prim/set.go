package prim

type Set[K comparable] map[K]struct{}

func (m *Set[K]) create() {
	if *m == nil {
		*m = make(Set[K])
	}
}

func (m *Set[K]) Add(keys ...K) {
	m.create()
	for _, key := range keys {
		(*m)[key] = struct{}{}
	}
}

func (m *Set[K]) Del(keys ...K) {
	if *m == nil {
		return
	}
	for _, key := range keys {
		delete(*m, key)
	}
}

func (m *Set[K]) Has(key K) bool {
	if *m == nil {
		return false
	}
	_, has := (*m)[key]
	return has
}

func (m *Set[K]) CopyFrom(other Set[K]) {
	if other == nil {
		return
	}
	m.create()
	for k := range other {
		(*m)[k] = struct{}{}
	}
}

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
