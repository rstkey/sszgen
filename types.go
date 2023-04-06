package main

import (
	"errors"
	"fmt"
	"go/types"
)

type SSZType interface {
	Fixed() bool
	FixedSize() int
	Name() string
}

func buildType(named *types.Named, typ types.Type) (SSZType, error) {
	switch t := typ.(type) {
	case *types.Named:
		if isBigInt(typ) {
			//return bigIntOp{}, nil
		}
		if isUint256(typ) {
			//return uint256Op{}, nil
		}
		return buildType(t, typ.Underlying())
	case *types.Basic:
		return newBasic(t, named)
	case *types.Array:
		return newVector(t, named)
	case *types.Slice:
		return newList(t, named)
	case *types.Pointer:
	case *types.Struct:
		return newStruct(t, named)
	case *types.Interface:
	}
	return nil, errors.New("unsupported type")
}

type Basic struct {
	types.Object
	basic *types.Basic
	fsize int
}

func newBasic(typ *types.Basic, named *types.Named) (*Basic, error) {
	var (
		size int
		kind = typ.Kind()
	)
	switch {
	case kind == types.Bool:
		size = 1
	case kind >= types.Uint8 && kind <= types.Uint64:
		size = 1 << (kind - types.Uint8)
	default:
		return nil, fmt.Errorf("unhandled basic type: %v", typ)
	}
	return &Basic{
		Object: named.Obj(),
		basic:  typ,
		fsize:  size,
	}, nil
}

func (b *Basic) Fixed() bool {
	return true
}

func (b *Basic) FixedSize() int {
	return b.fsize
}

type Vector struct {
	types.Object
	*types.Array
	elem SSZType
}

func newVector(typ *types.Array, named *types.Named) (*Vector, error) {
	elem, err := buildType(nil, typ.Elem())
	if err != nil {
		return nil, err
	}
	return &Vector{
		Object: named.Obj(),
		Array:  typ,
		elem:   elem,
	}, nil
}

func (v *Vector) Fixed() bool {
	return v.elem.Fixed()
}

func (v *Vector) FixedSize() int {
	if v.Fixed() {
		return int(v.Len()) * v.elem.FixedSize()
	}
	return BytesPerLengthOffset
}

type List struct {
	types.Object
	*types.Slice
	elem SSZType
}

func newList(typ *types.Slice, named *types.Named) (*List, error) {
	elem, err := buildType(nil, typ.Elem())
	if err != nil {
		return nil, err
	}
	return &List{
		Object: named.Obj(),
		Slice:  typ,
		elem:   elem,
	}, nil
}

func (l *List) Fixed() bool {
	return false
}

func (l *List) FixedSize() int {
	return BytesPerLengthOffset
}

type Struct struct {
	types.Object
	*types.Struct
	fields []SSZType
}

func newStruct(typ *types.Struct, named *types.Named) (*Struct, error) {
	var fields []SSZType
	for i := 0; i < typ.NumFields(); i++ {
		field, err := buildType(nil, typ.Field(i).Type())
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}
	return &Struct{
		Object: named.Obj(),
		Struct: typ,
		fields: fields,
	}, nil
}

func (s *Struct) Fixed() bool {
	for _, field := range s.fields {
		if !field.Fixed() {
			return false
		}
	}
	return true
}

func (s *Struct) FixedSize() int {
	if !s.Fixed() {
		return BytesPerLengthOffset
	}
	var size int
	for _, field := range s.fields {
		size += field.FixedSize()
	}
	return size
}

// isBigInt checks whether 'typ' is "math/big".Int.
func isBigInt(typ types.Type) bool {
	named, ok := typ.(*types.Named)
	if !ok {
		return false
	}
	name := named.Obj()
	return name.Pkg().Path() == "math/big" && name.Name() == "Int"
}

// isUint256 checks whether 'typ' is "github.com/holiman/uint256".Int.
func isUint256(typ types.Type) bool {
	named, ok := typ.(*types.Named)
	if !ok {
		return false
	}
	name := named.Obj()
	return name.Pkg().Path() == "github.com/holiman/uint256" && name.Name() == "Int"
}

// isByte checks whether the underlying type of 'typ' is uint8.
func isByte(typ types.Type) bool {
	basic, ok := resolveUnderlying(typ).(*types.Basic)
	return ok && basic.Kind() == types.Uint8
}

func resolveUnderlying(typ types.Type) types.Type {
	for {
		t := typ.Underlying()
		if t == typ {
			return t
		}
		typ = t
	}
}
