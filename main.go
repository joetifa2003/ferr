package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"strings"
)

func main() {
	fset := token.NewFileSet()
	pkgsMap, err := parser.ParseDir(fset, "pkg", nil, parser.SkipObjectResolution)
	if err != nil {
		panic(err)
	}

	pkgs := MapValues(pkgsMap)

	filesMap := map[string]*ast.File{}
	for _, p := range pkgs {
		for k, v := range p.Files {
			filesMap[k] = v
		}
	}

	files := MapValues(filesMap)

	pos := 8
	file := "pkg/test.go"

	info := types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	var conf types.Config
	conf.Importer = importer.ForCompiler(fset, "gc", func(path string) (io.ReadCloser, error) {
		return nil, nil
	})

	_, err = conf.Check("", fset, files, &info)
	if err != nil {
		panic(err)
	}

	var typeList *ast.FieldList
	ast.Inspect(filesMap[file], func(n ast.Node) bool {
		if n == nil {
			return false
		}

		lineStart := fset.Position(n.Pos()).Line
		lineEnd := fset.Position(n.End()).Line
		if pos < lineStart || pos > lineEnd {
			return false
		}

		switch n := n.(type) {
		case *ast.FuncLit:
			typeList = n.Type.Results
		case *ast.FuncDecl:
			typeList = n.Type.Results
		}

		return true
	})

	if typeList == nil {
		fmt.Println("not in a function")
		return
	}

	finalTypes := Map(typeList.List, func(f *ast.Field) string { return getFinalType(info, f.Type).zeroExpr() })

	for _, t := range finalTypes {
		fmt.Println(t)
	}
}

func Map[T, U any](arr []T, f func(T) U) []U {
	res := make([]U, 0, len(arr))

	for _, v := range arr {
		res = append(res, f(v))
	}

	return res
}

func MapValues[T comparable, U any](m map[T]U) []U {
	res := make([]U, 0, len(m))

	for _, v := range m {
		res = append(res, v)
	}

	return res
}

type finalType struct {
	value string
	name  string
}

func (f finalType) zeroExpr() string {
	if strings.HasPrefix(f.value, "*") || strings.HasPrefix(f.value, "map[") || strings.HasPrefix(f.value, "interface{") {
		return "nil"
	}

	if strings.HasPrefix(f.value, "struct{") {
		return fmt.Sprintf("%s{}", f.name)
	}

	if f.name == "error" {
		return "err"
	}

	panic(fmt.Sprintf("cannot find zero value of %s", f))
}

func getFinalType(info types.Info, expr ast.Expr) finalType {
	name := ""
	if ident, ok := expr.(*ast.Ident); ok {
		name = ident.Name
	}

	t := info.TypeOf(expr)
	for {
		if t.String() == t.Underlying().String() {
			return finalType{
				value: t.String(),
				name:  name,
			}
		}

		t = t.Underlying()
	}
}

func generateIfErr(typs []string) string {
	returnExpr := strings.Join(typs, ", ")

	return fmt.Sprintf("if err != nil {\n  %s\n}", returnExpr)
}
