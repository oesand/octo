package mediator

import (
	"testing"

	"github.com/oesand/octo"
)

func TestInject_Singleton(t *testing.T) {
	container := octo.New()
	m1 := Inject(container)
	m2 := Inject(container)

	if m1 != m2 {
		t.Fatal("expected same manager instance")
	}

	container = octo.New()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when injecting manually created Manager")
		}
	}()

	octo.InjectValue(container, &Manager{})

	Inject(container)
}
