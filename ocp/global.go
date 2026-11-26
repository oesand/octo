package ocp

import (
	"github.com/oesand/octo"
	"github.com/oesand/octo/internal"
)

type Injector func(*octo.Container)

var groupedPack = map[string][]Injector{}

func ResolveNamed[T any](name string) octo.Provider[T] {
	return func(container *octo.Container) T {
		return octo.ResolveNamed[T](container, name)
	}
}

func Collect(group string, injectors ...Injector) internal.Token {
	pack, has := groupedPack[group]
	if !has {
		pack = []Injector{}
	}

	pack = append(pack, injectors...)
	groupedPack[group] = pack

	return internal.Token{}
}

func InjectGroup(container *octo.Container, group string) int {
	pack, has := groupedPack[group]
	if !has {
		return 0
	}

	for _, injector := range pack {
		injector(container)
	}

	return len(pack)
}
