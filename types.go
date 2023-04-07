package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/types"

	"github.com/rjl493456442/sszgen/ssz"
)

type sszType interface {
	fixed() bool
	fixedSize() int
	typeName() string
	genSize(ctx *genContext, w string, obj string) string
	genEncoder(ctx *genContext, obj string) string
}

func buildType(named *types.Named, typ types.Type, tags []sizeTag) (sszType, error) {
	switch t := typ.(type) {
	case *types.Named:
		if isBigInt(typ) {
			//return bigIntOp{}, nil
		}
		if isUint256(typ) {
			//return uint256Op{}, nil
		}
		return buildType(t, typ.Underlying(), tags)
	case *types.Basic:
		return newBasic(named, t)
	case *types.Array:
		return newVector(named, t, tags)
	case *types.Slice:
		return newList(named, t, tags)
	case *types.Pointer:
		if isBigInt(t.Elem()) {
			//return bigIntOp{pointer: true}, nil
		}
		if isUint256(t.Elem()) {
			//return uint256Op{pointer: true}, nil
		}
		return newPointer(named, t, tags)
	case *types.Struct:
		return newStruct(named, t)
	case *types.Interface:
	}
	return nil, errors.New("unsupported type")
}

type Basic struct {
	basic   *types.Basic
	named   *types.Named
	size    int
	encoder string
}

func newBasic(named *types.Named, typ *types.Basic) (*Basic, error) {
	var (
		size    int
		encoder string
		kind    = typ.Kind()
	)
	switch {
	case kind == types.Bool:
		size = 1
		encoder = "EncodeBool"
	case kind == types.Uint8:
		encoder = "EncodeByte"
	case kind > types.Uint8 && kind <= types.Uint64:
		size = 1 << (kind - types.Uint8)
		encoder = fmt.Sprintf("EncodeUint%d", size*8)
	default:
		return nil, fmt.Errorf("unhandled basic type: %v", typ)
	}
	return &Basic{
		basic:   typ,
		named:   named,
		size:    size,
		encoder: encoder,
	}, nil
}

func (b *Basic) fixed() bool {
	return true
}

func (b *Basic) fixedSize() int {
	return b.size
}

func (b *Basic) typeName() string {
	return b.basic.String()
}

func (b *Basic) genSize(ctx *genContext, w string, obj string) string {
	return fmt.Sprintf("%s += %d\n", w, b.size)
}

func (b *Basic) genEncoder(ctx *genContext, obj string) string {
	ctx.addImport(pkgPath, "")
	if b.named != nil {
		obj = fmt.Sprintf("uint%d(%s)", b.size*8, obj)
	}
	return fmt.Sprintf("%s(w, %s)\n", ctx.qualifier(pkgPath, b.encoder), obj)
}

type Vector struct {
	array   *types.Array
	named   *types.Named
	elem    sszType
	len     int64
	tag     sizeTag
	encoder string
}

func newVector(named *types.Named, typ *types.Array, tags []sizeTag) (*Vector, error) {
	var (
		tag    sizeTag
		remain []sizeTag
	)
	if len(tags) > 0 {
		tag, remain = tags[0], tags[1:]
	}
	if tag.size != 0 && tag.size != typ.Len() {
		return nil, fmt.Errorf("invalid size tag, array: %d, tag %d", typ.Len(), tag.size)
	}
	if tag.limit != 0 {
		return nil, fmt.Errorf("unexpected size limit tag")
	}
	elem, err := buildType(nil, typ.Elem(), remain)
	if err != nil {
		return nil, err
	}
	var encoder string
	if b, ok := elem.(*Basic); ok {
		encoder = fmt.Sprintf("%ss", b.encoder)
	}
	return &Vector{
		array:   typ,
		named:   named,
		elem:    elem,
		len:     typ.Len(),
		tag:     tag,
		encoder: encoder,
	}, nil
}

func (v *Vector) fixed() bool {
	return v.elem.fixed()
}

func (v *Vector) fixedSize() int {
	if v.fixed() {
		return int(v.len) * v.elem.fixedSize()
	}
	return ssz.BytesPerLengthOffset
}

func (v *Vector) typeName() string {
	return v.array.String()
}

func (v *Vector) genSize(ctx *genContext, w string, obj string) string {
	if v.elem.fixed() {
		return fmt.Sprintf("%s += %d\n", w, int(v.len)*v.elem.fixedSize())
	}
	var (
		b  bytes.Buffer
		vn = ctx.tmpVar()
	)
	fmt.Fprintf(&b, "for _, %s := range %s {\n", vn, obj)
	fmt.Fprintf(&b, "%s", v.elem.genSize(ctx, w, vn))
	fmt.Fprint(&b, "}\n")
	return b.String()
}

func (v *Vector) genEncoder(ctx *genContext, obj string) string {
	if v.encoder != "" {
		return fmt.Sprintf("%s(w, %s[:])\n", ctx.qualifier(pkgPath, v.encoder), obj)
	}
	var b bytes.Buffer
	if !v.elem.fixed() {
		fmt.Fprintf(&b, "_o = len(%s)*4\n", obj)

		vn := ctx.tmpVar()
		fmt.Fprintf(&b, "for _, %s := range %s {\n", vn, obj)
		fmt.Fprintf(&b, "%s(w, uint32(_o))\n", ctx.qualifier(pkgPath, "EncodeUint32"))
		fmt.Fprintf(&b, "%s", v.elem.genSize(ctx, "_o", vn))
		fmt.Fprint(&b, "}\n")
	}
	vn := ctx.tmpVar()
	fmt.Fprintf(&b, "for _, %s := range %s {\n", vn, obj)
	fmt.Fprintf(&b, "%s", v.elem.genEncoder(ctx, vn))
	fmt.Fprint(&b, "}\n")
	return b.String()
}

type List struct {
	slice   *types.Slice
	named   *types.Named
	elem    sszType
	tag     sizeTag
	encoder string
}

func newList(named *types.Named, slice *types.Slice, tags []sizeTag) (*List, error) {
	var (
		tag    sizeTag
		remain []sizeTag
	)
	if len(tags) > 0 {
		tag, remain = tags[0], tags[1:]
	}
	elem, err := buildType(nil, slice.Elem(), remain)
	if err != nil {
		return nil, err
	}
	var encoder string
	if b, ok := elem.(*Basic); ok {
		encoder = fmt.Sprintf("%ss", b.encoder)
	}
	return &List{
		slice:   slice,
		named:   named,
		elem:    elem,
		tag:     tag,
		encoder: encoder,
	}, nil
}

func (l *List) fixed() bool {
	if l.tag.size != 0 {
		return l.elem.fixed()
	}
	return false
}

func (l *List) fixedSize() int {
	if l.fixed() {
		return int(l.tag.size) * l.elem.fixedSize()
	}
	return ssz.BytesPerLengthOffset
}

func (l *List) typeName() string {
	return l.slice.String()
}

func (l *List) genSize(ctx *genContext, w string, obj string) string {
	if l.elem.fixed() {
		if l.elem.fixedSize() == 1 {
			return fmt.Sprintf("%s += len(%s)\n", w, obj)
		}
		return fmt.Sprintf("%s += len(%s)*%d\n", w, obj, l.elem.fixedSize())
	}
	var (
		b  bytes.Buffer
		vn = ctx.tmpVar()
	)
	fmt.Fprintf(&b, "for _, %s := range %s {\n", vn, obj)
	fmt.Fprintf(&b, "%s += 4\n", w)
	fmt.Fprintf(&b, "%s", l.elem.genSize(ctx, w, vn))
	fmt.Fprint(&b, "}\n")
	return b.String()
}

func (l *List) genEncoder(ctx *genContext, obj string) string {
	if l.encoder != "" {
		return fmt.Sprintf("%s(w, %s)\n", ctx.qualifier(pkgPath, l.encoder), obj)
	}
	var b bytes.Buffer
	if !l.elem.fixed() {
		fmt.Fprintf(&b, "_o = len(%s)*4\n", obj)

		vn := ctx.tmpVar()
		fmt.Fprintf(&b, "for _, %s := range %s {\n", vn, obj)
		fmt.Fprintf(&b, "%s(w, uint32(_o))\n", ctx.qualifier(pkgPath, "EncodeUint32"))
		fmt.Fprintf(&b, "%s", l.elem.genSize(ctx, "_o", vn))
		fmt.Fprint(&b, "}\n")
	}
	vn := ctx.tmpVar()
	fmt.Fprintf(&b, "for _, %s := range %s {\n", vn, obj)
	fmt.Fprintf(&b, "%s", l.elem.genEncoder(ctx, vn))
	fmt.Fprint(&b, "}\n")
	return b.String()
}

type Struct struct {
	*types.Struct
	named      *types.Named
	fields     []sszType
	fieldNames []string
}

func newStruct(named *types.Named, typ *types.Struct) (*Struct, error) {
	var (
		fields     []sszType
		fieldNames []string
	)
	for i := 0; i < typ.NumFields(); i++ {
		f := typ.Field(i)
		if !f.Exported() {
			continue
		}
		var (
			err     error
			ignored bool
			tags    []sizeTag
		)
		if tag := typ.Tag(i); tag != "" {
			ignored, tags, err = parseTag(tag)
			if err != nil {
				return nil, err
			}
		}
		if ignored {
			continue
		}
		field, err := buildType(nil, f.Type(), tags)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
		fieldNames = append(fieldNames, f.Name())
	}
	return &Struct{
		Struct:     typ,
		named:      named,
		fields:     fields,
		fieldNames: fieldNames,
	}, nil
}

func (s *Struct) fixed() bool {
	for _, field := range s.fields {
		if !field.fixed() {
			return false
		}
	}
	return true
}

func (s *Struct) fixedSize() int {
	if !s.fixed() {
		return ssz.BytesPerLengthOffset
	}
	var size int
	for _, field := range s.fields {
		size += field.fixedSize()
	}
	return size
}

func (s *Struct) genSize(ctx *genContext, w string, obj string) string {
	if !ctx.topType {
		return fmt.Sprintf("%s += %s.SizeSSZ()\n", w, obj)
	}
	ctx.topType = false

	var b bytes.Buffer
	var fixedSize int
	for _, field := range s.fields {
		fixedSize += field.fixedSize()
	}
	fmt.Fprintf(&b, "%s := %d\n", w, fixedSize)

	for i, field := range s.fields {
		if field.fixed() {
			continue
		}
		fmt.Fprintf(&b, "%s", field.genSize(ctx, w, fmt.Sprintf("%s.%s", obj, s.fieldNames[i])))
	}
	return b.String()
}

func (s *Struct) genEncoder(ctx *genContext, obj string) string {
	var b bytes.Buffer
	if !ctx.topType {
		fmt.Fprintf(&b, "if err := %s.MarshalSSZTo(w); err != nil {\n", obj)
		fmt.Fprint(&b, "return err\n")
		fmt.Fprint(&b, "}\n")
		return b.String()
	}
	ctx.topType = false

	if !s.fixed() {
		var offset int
		for _, field := range s.fields {
			offset += field.fixedSize()
		}
		fmt.Fprintf(&b, "_o := %d\n", offset)
	}
	for i, field := range s.fields {
		if field.fixed() {
			fmt.Fprintf(&b, "%s", field.genEncoder(ctx, fmt.Sprintf("%s.%s", obj, s.fieldNames[i])))
		} else {
			fmt.Fprintf(&b, "%s(w, uint32(_o))\n", ctx.qualifier(pkgPath, "EncodeUint32"))
			fmt.Fprintf(&b, "%s", field.genSize(ctx, "_o", fmt.Sprintf("%s.%s", obj, s.fieldNames[i])))
		}
	}
	for i, field := range s.fields {
		if field.fixed() {
			continue
		}
		fmt.Fprintf(&b, "%s", field.genEncoder(ctx, fmt.Sprintf("%s.%s", obj, s.fieldNames[i])))
	}
	return b.String()
}

func (s *Struct) typeName() string {
	return s.named.Obj().Name()
}

type Pointer struct {
	*types.Pointer
	named *types.Named
	elem  sszType
}

func newPointer(named *types.Named, typ *types.Pointer, tags []sizeTag) (*Pointer, error) {
	elem, err := buildType(nil, typ.Elem(), tags)
	if err != nil {
		return nil, err
	}
	return &Pointer{
		Pointer: typ,
		named:   named,
		elem:    elem,
	}, nil
}

func (p *Pointer) typeName() string {
	return fmt.Sprintf("*%s", p.elem.typeName())
}

func (p *Pointer) fixed() bool {
	return p.elem.fixed()
}

func (p *Pointer) fixedSize() int {
	return p.elem.fixedSize()
}

func (p *Pointer) genSize(ctx *genContext, w string, obj string) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "if %s == nil {\n", obj)
	fmt.Fprintf(&b, "%s = new(%s)\n", obj, p.elem.typeName())
	fmt.Fprint(&b, "}\n")
	fmt.Fprintf(&b, "%s", p.elem.genSize(ctx, w, obj))
	return b.String()
}

func (p *Pointer) genEncoder(ctx *genContext, obj string) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "if %s == nil {\n", obj)
	fmt.Fprintf(&b, "%s = new(%s)\n", obj, p.elem.typeName())
	fmt.Fprint(&b, "}\n")
	fmt.Fprintf(&b, "%s", p.elem.genEncoder(ctx, obj))
	return b.String()
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
