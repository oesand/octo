package parse

import (
	"go/types"
)

func implementsMediatrHandlers(named *types.Named) bool {
	for fun := range named.Methods() {
		switch fun.Name() {
		case "Notification":
			sig := fun.Signature()
			return sig.Params().Len() == 2 &&
				sig.Results().Len() == 0 &&
				isContextType(sig.Params().At(0).Type())

		case "Request":
			sig := fun.Signature()
			return sig.Params().Len() == 2 &&
				sig.Results().Len() == 2 &&
				isContextType(sig.Params().At(0).Type()) &&
				isErrorType(sig.Results().At(1).Type())
		}
	}

	return false
}

func isContextType(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	if named.Obj().Pkg() == nil {
		return false
	}
	return named.Obj().Pkg().Path() == "context" && named.Obj().Name() == "Context"
}

func isErrorType(t types.Type) bool {
	iface, ok := t.Underlying().(*types.Interface)
	if !ok {
		return false
	}
	return iface.NumMethods() == 1 && iface.Method(0).Name() == "Error"
}
