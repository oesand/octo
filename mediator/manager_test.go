package mediator_test

import (
	"testing"

	"github.com/oesand/octo"
	"github.com/oesand/octo/mediator"
)

func TestInject_Singleton(t *testing.T) {
	container := octo.New()
	m1 := mediator.Inject(container)
	m2 := mediator.Inject(container)

	if m1 != m2 {
		t.Fatal("expected same manager instance")
	}
}

func TestInject_ManualInject(t *testing.T) {
	container := octo.New()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when injecting manually created Manager")
		}
	}()

	octo.InjectValue(container, &mediator.Manager{})

	mediator.Inject(container)
}
