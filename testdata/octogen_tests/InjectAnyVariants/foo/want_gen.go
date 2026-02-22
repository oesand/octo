package foo

import (
	"github.com/oesand/octo"
	fnc "github.com/oesand/octo/testdata/octogen_tests/InjectAnyVariants/foo/fnc"
)

func IncludeFunc(container *octo.Container) {
	octo.Inject(container, func(container *octo.Container) *fnc.Struct {
		return fnc.NewPtrStruct(
			octo.Resolve[*fnc.Linked](container),
		)
	})
	octo.Inject(container, func(container *octo.Container) fnc.Struct {
		return fnc.NewStruct(
			octo.Resolve[fnc.Linked](container),
		)
	})
	octo.Inject(container, func(container *octo.Container) fnc.Iface {
		return fnc.NewIface(
			octo.Resolve[*fnc.Linked](container),
		)
	})
	octo.InjectNamed(container, "named", func(container *octo.Container) fnc.Iface {
		return fnc.NewIface(
			octo.Resolve[*fnc.Linked](container),
		)
	})
}
