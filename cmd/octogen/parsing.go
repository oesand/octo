//go:build octogen
// +build octogen

package main

import (
	gg "github.com/oesand/octo/octogen"
)

type Inf interface{}

type Other struct{}

type Named struct {
	Other *Other
}

type Struct struct {
	Named *Named `key:"key1"`
	Other *Other
}

func (s *Struct) Do() {

}

func NewHelloWorld(p []Inf) *Other {
	return &Other{}
}

func Include() {
	gg.Inject[*Other]()
	gg.Inject[*Named]("key1")
	gg.Inject[*Struct]()

	gg.Inject(NewHelloWorld)
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
