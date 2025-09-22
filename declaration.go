package octo

import (
	"reflect"
	"sync"
)

func newDeclLazy(typ reflect.Type, name string, provider func() any) ServiceDeclaration {
	return &serviceDeclLazy{
		rtyp:     typ,
		name:     name,
		provider: provider,
	}
}

type serviceDeclLazy struct {
	mu sync.Mutex

	name     string
	rtyp     reflect.Type
	built    bool
	instance any
	provider func() any
}

func (decl *serviceDeclLazy) Name() string {
	return decl.name
}

func (decl *serviceDeclLazy) Value() any {
	decl.mu.Lock()
	defer decl.mu.Unlock()

	if decl.built {
		return decl.instance
	}

	val := decl.provider()
	decl.built = true
	decl.instance = val

	return val
}

func (decl *serviceDeclLazy) Type() reflect.Type {
	return decl.rtyp
}

func newDeclValue(typ reflect.Type, name string, instance any) ServiceDeclaration {
	return &serviceDeclValue{
		rtyp:     typ,
		name:     name,
		instance: instance,
	}
}

type serviceDeclValue struct {
	name     string
	rtyp     reflect.Type
	instance any
}

func (decl *serviceDeclValue) Name() string {
	return decl.name
}

func (decl *serviceDeclValue) Value() any {
	return decl.instance
}

func (decl *serviceDeclValue) Type() reflect.Type {
	return decl.rtyp
}
