package main

import (
	"errors"
	"go/types"
)

func parsePackage(pkg *types.Package, names []string) ([]SSZType, error) {
	if len(names) == 0 {
		names = pkg.Scope().Names()
	}
	var types []SSZType
	for _, name := range names {
		named, err := lookupStructType(pkg.Scope(), name)
		if err != nil {
			return nil, err
		}
		typ, err := buildType(nil, named)
		if err != nil {
			return nil, err
		}
		types = append(types, typ)
	}
	return types, nil
}

func lookupStructType(scope *types.Scope, name string) (*types.Named, error) {
	typ, err := lookupType(scope, name)
	if err != nil {
		return nil, err
	}
	_, ok := typ.Underlying().(*types.Struct)
	if !ok {
		return nil, errors.New("not a struct type")
	}
	return typ, nil
}

func lookupType(scope *types.Scope, name string) (*types.Named, error) {
	obj := scope.Lookup(name)
	if obj == nil {
		return nil, errors.New("no such identifier")
	}
	typ, ok := obj.(*types.TypeName)
	if !ok {
		return nil, errors.New("not a type")
	}
	return typ.Type().(*types.Named), nil
}
