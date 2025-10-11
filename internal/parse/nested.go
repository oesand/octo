package parse

import (
	"errors"
	"github.com/oesand/octo/internal/decl"
	"go/types"
)

type ParseExtraOption int

const (
	NoEP    ParseExtraOption = 0
	EPSlice ParseExtraOption = 1 << iota
	EPIface
)

func unexpectedTypeErr(options ParseExtraOption) error {
	errText := "unexpected type, supported only struct, pointer struct"

	if options&EPIface != 0 {
		errText += ", interface"
	}

	if options&EPSlice != 0 {
		errText += " and slice of them"
	}

	return errors.New(errText)
}

func parseTypeLink(typ types.Type, options ParseExtraOption) (types.Type, *decl.LocaleInfo, error) {
	sl, sliced := typ.(*types.Slice)
	if sliced {
		if options&EPSlice == 0 {
			return nil, nil, unexpectedTypeErr(options)
		}

		typ = sl.Elem()
	}

	pointer, ptr := typ.(*types.Pointer)
	if ptr {
		typ = pointer.Elem()
	}

	named, ok := typ.(*types.Named)
	if !ok {
		return nil, nil, unexpectedTypeErr(options)
	}

	if named.TypeParams().Len() > 0 {
		return nil, nil, errors.New("types with generics not supported")
	}

	typ = named.Underlying()
	switch typ.(type) {
	case *types.Struct:
	case *types.Interface:
		if ptr || options&EPIface == 0 {
			return nil, nil, unexpectedTypeErr(options)
		}
	default:
		return nil, nil, unexpectedTypeErr(options)
	}

	loc := decl.LocaleInfo{
		Sliced:  sliced,
		Ptr:     ptr,
		Name:    named.Obj().Name(),
		Package: named.Obj().Pkg().Path(),
	}
	return typ, &loc, nil
}
