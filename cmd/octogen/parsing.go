//go:build octogen
// +build octogen

package main

import gg "github.com/oesand/octo/octogen"

type Other struct{}

type Named struct {
	Other *Other
}

type Struct struct {
	Named *Named `key:"key1"`
	Other *Other
}

func Include() {
	gg.Inject[*Other]()
	gg.InjectNamed[*Named]("key1")
	gg.Inject[*Struct]()
}

/*

func Include(container *octo.Container) {
	octo.Inject(container, func(container *octo.Container) *Other {
		return &Other{}
	})

	octo.InjectNamed(container, "key1", func(container *octo.Container) *Named {
		return &Named{
			Other: octo.Resolve[*Other](container),
		}
	})

	octo.Inject(container, func(container *octo.Container) *Struct {
		return &Struct{
			Other: octo.Resolve[*Other](container),
			Named: octo.ResolveNamed[*Named](container, "key1"),
		}
	})
}

*/
