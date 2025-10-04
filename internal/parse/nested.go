package parse

import (
	"errors"
	"github.com/oesand/octo/internal/decl"
	"go/types"
)

func parseStructLocale(typ types.Type) (*types.Struct, *decl.LocaleInfo, error) {
	const unexpectedTypeErr = "unexpected type, supported only struct, pointer struct"

	pointer, ptr := typ.(*types.Pointer)
	if ptr {
		typ = pointer.Elem()
	}

	named, ok := typ.(*types.Named)
	if !ok {
		return nil, nil, errors.New(unexpectedTypeErr)
	}

	stct, ok := named.Underlying().(*types.Struct)
	if !ok {
		return nil, nil, errors.New(unexpectedTypeErr)
	}

	loc := decl.LocaleInfo{
		Ptr:     ptr,
		Name:    named.Obj().Name(),
		Package: named.Obj().Pkg().Path(),
	}
	return stct, &loc, nil
}

func parseFieldLocale(typ types.Type) (*decl.LocaleInfo, error) {
	const unexpectedTypeErr = "unexpected type, supported only struct, pointer struct and interface and slice of them"

	sl, sliced := typ.(*types.Slice)
	if sliced {
		typ = sl.Elem()
	}

	pointer, ptr := typ.(*types.Pointer)
	if ptr {
		typ = pointer.Elem()
	}

	named, ok := typ.(*types.Named)
	if !ok {
		return nil, errors.New(unexpectedTypeErr)
	}

	switch named.Underlying().(type) {
	case *types.Struct:
	case *types.Interface:
		if ptr {
			return nil, errors.New(unexpectedTypeErr)
		}
	default:
		return nil, errors.New(unexpectedTypeErr)
	}

	loc := decl.LocaleInfo{
		Sliced:  sliced,
		Ptr:     ptr,
		Name:    named.Obj().Name(),
		Package: named.Obj().Pkg().Path(),
	}
	return &loc, nil
}
