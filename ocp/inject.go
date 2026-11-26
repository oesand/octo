package ocp

import "github.com/oesand/octo"

func Inject[R any](provider func() R) Injector {
	return func(container *octo.Container) {
		octo.Inject(container, func(container *octo.Container) R {
			return provider()
		})
	}
}

func Inject1[P1 any, R any](
	getter1 octo.Provider[P1],
	provider func(P1) R,
) Injector {
	return func(container *octo.Container) {
		octo.Inject(container, func(container *octo.Container) R {
			return provider(getter1(container))
		})
	}
}

func InjectR1[P1 any, R any](
	provider func(P1) R,
) Injector {
	return func(container *octo.Container) {
		octo.Inject(container, func(container *octo.Container) R {
			return provider(octo.Resolve[P1](container))
		})
	}
}

func Inject2[P1 any, P2 any, R any](
	getter1 octo.Provider[P1],
	getter2 octo.Provider[P2],
	provider func(P1, P2) R,
) Injector {
	return func(container *octo.Container) {
		octo.Inject(container, func(container *octo.Container) R {
			return provider(
				getter1(container),
				getter2(container),
			)
		})
	}
}

func InjectR2[P1 any, P2 any, R any](
	provider func(P1, P2) R,
) Injector {
	return func(container *octo.Container) {
		octo.Inject(container, func(container *octo.Container) R {
			return provider(
				octo.Resolve[P1](container),
				octo.Resolve[P2](container),
			)
		})
	}
}
