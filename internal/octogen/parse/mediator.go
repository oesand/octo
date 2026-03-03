package parse

/*
func scanForMediator(pkg packages.Package) {
	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		structType := scope.Lookup(name).Type()

		_, named, _, ok := splitStructType(structType)
		if !ok || !isImplementsMediatrHandlers(named){
			continue
		}

		if funcObj := scope.Lookup(name); funcObj != nil {

		}

	}
}


func isImplementsMediatrHandlers(named *types.Named) bool {
	for fun := range named.Methods() {
		switch fun.Name() {
		case "Notification":
			sig := fun.Signature()
			return sig.Params().Len() == 2 &&
				sig.Results().Len() == 1 &&
				isContextType(sig.Params().At(0).Type()) &&
				isErrorType(sig.Results().At(0).Type())

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


*/
