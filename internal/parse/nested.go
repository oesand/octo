package parse

import "go/types"

func parseTypeLocale(typ types.Type) (types.Type, *LocaleInfo) {
	var ptrLevel int
	for {
		if t, ok := typ.(*types.Pointer); ok {
			typ = t.Elem()
			ptrLevel++
		} else {
			break
		}
	}

	named, ok := typ.(*types.Named)
	if !ok {
		return nil, nil
	}

	obj := named.Obj()
	locale := LocaleInfo{
		PtrLevel: ptrLevel,
		Name:     obj.Name(),
		Package:  obj.Pkg().Path(),
	}

	return named.Underlying(), &locale
}
