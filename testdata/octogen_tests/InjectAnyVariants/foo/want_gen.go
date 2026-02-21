package foo

import (
	"github.com/oesand/octo"
)

func IncludeAny(container *octo.Container) {
	octo.Inject(container, func(container *octo.Container) *NewestStruct {
		return NewStruct(
			octo.Resolve[Inf](container),
			octo.ResolveAll[Inf](container),
			octo.Resolve[*Other](container),
			octo.Resolve[Struct](container),
			octo.Resolve[*Named](container),
		)
	})
	octo.Inject(container, func(container *octo.Container) NewestStruct {
		return NewStct(
			octo.Resolve[Inf](container),
			octo.ResolveAll[Inf](container),
			octo.Resolve[Other](container),
			octo.Resolve[*Struct](container),
			octo.Resolve[Named](container),
		)
	})
	octo.InjectNamed(container, "key2", func(container *octo.Container) NewestStruct {
		return NewStct(
			octo.Resolve[Inf](container),
			octo.ResolveAll[Inf](container),
			octo.Resolve[Other](container),
			octo.Resolve[*Struct](container),
			octo.Resolve[Named](container),
		)
	})
	octo.Inject(container, func(container *octo.Container) Inf {
		return NewInf(
			octo.Resolve[*Struct](container),
		)
	})
}
