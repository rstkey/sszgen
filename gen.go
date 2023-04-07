package main

import (
	"bytes"
	"fmt"
	"go/types"
	"sort"
)

const pkgPath = "github.com/rjl493456442/sszgen/ssz"

type genContext struct {
	topType bool
	pkg     *types.Package
	imports map[string]string
	vid     int
}

func newGenContext(pkg *types.Package) *genContext {
	return &genContext{
		pkg:     pkg,
		imports: make(map[string]string),
	}
}

func (ctx *genContext) qualifier(path string, obj string) string {
	if path == ctx.pkg.Path() {
		return obj
	}
	return fmt.Sprintf("%s.%s", pkgName(path), obj)
}

func (ctx *genContext) addImport(path string, alias string) error {
	if path == ctx.pkg.Path() {
		return nil
	}
	if n, ok := ctx.imports[path]; n != alias && ok {
		return fmt.Errorf("conflict import %s(alias: %s-%s)", path, n, alias)
	}
	ctx.imports[path] = alias
	return nil
}

func (ctx *genContext) header() []byte {
	var paths sort.StringSlice
	for path := range ctx.imports {
		paths = append(paths, path)
	}
	sort.Sort(paths)

	var b bytes.Buffer
	fmt.Fprintf(&b, "package %s\n", ctx.pkg.Name())
	if len(paths) == 0 {
		return b.Bytes()
	}
	if len(paths) == 1 {
		alias := ctx.imports[paths[0]]
		if alias == "" {
			fmt.Fprintf(&b, "import \"%s\"\n", paths[0])
		} else {
			fmt.Fprintf(&b, "import %s \"%s\"\n", alias, paths[0])
		}
		return b.Bytes()
	}
	fmt.Fprintf(&b, "import (\n")
	for _, path := range paths {
		alias := ctx.imports[path]
		if alias == "" {
			fmt.Fprintf(&b, "\"%s\"\n", path)
		} else {
			fmt.Fprintf(&b, "%s \"%s\"\n", alias, path)
		}
	}
	fmt.Fprintf(&b, ")\n")
	return b.Bytes()
}

func (ctx *genContext) tmpVar() string {
	id := fmt.Sprintf("_v%d", ctx.vid)
	ctx.vid += 1
	return id
}

func (ctx *genContext) reset() {
	ctx.vid = 0
	ctx.topType = true
}

func generateSSZSize(ctx *genContext, typ sszType) ([]byte, error) {
	var b bytes.Buffer
	ctx.reset()

	// TODO non-struct types are not supported yet
	if _, ok := typ.(*Struct); !ok {
		return nil, nil
	}
	fmt.Fprintf(&b, "func (obj *%s) SizeSSZ() int {\n", typ.typeName())
	fmt.Fprint(&b, typ.genSize(ctx, "size", "obj"))
	fmt.Fprint(&b, "return size\n")
	fmt.Fprintf(&b, "}\n")
	return b.Bytes(), nil
}

func generateEncoder(ctx *genContext, typ sszType) ([]byte, error) {
	var b bytes.Buffer
	ctx.reset()

	// TODO non-struct types are not supported yet
	if _, ok := typ.(*Struct); !ok {
		return nil, nil
	}

	// Generate `MarshalSSZTo` binding
	fmt.Fprintf(&b, "func (obj *%s) MarshalSSZTo(w []byte) error {\n", typ.typeName())
	fmt.Fprint(&b, typ.genEncoder(ctx, "obj"))
	fmt.Fprint(&b, "return nil\n")
	fmt.Fprint(&b, "}\n")
	return b.Bytes(), nil
}

func generate(ctx *genContext, typ sszType) ([]byte, error) {
	var codes [][]byte
	for _, fn := range []func(ctx *genContext, typ sszType) ([]byte, error){
		generateSSZSize,
		generateEncoder,
	} {
		code, err := fn(ctx, typ)
		if err != nil {
			return nil, err
		}
		codes = append(codes, code)
	}
	//fmt.Println(string(bytes.Join(codes, []byte("\n"))))
	return bytes.Join(codes, []byte("\n")), nil
}
