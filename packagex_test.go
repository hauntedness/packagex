package packagex

import (
	"go/ast"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestVisitor_Walk(t *testing.T) {
	v := Visitor{}
	v.StructVisitor = func(st *ast.StructType, _ *ast.GenDecl, _ *ast.File, _ *packages.Package) error {
		for _, field := range st.Fields.List {
			//fmt.Println(field.Names[0].Name)
			if st, ok := field.Type.(*ast.StructType); ok && len(field.Names) == 1 && field.Names[0].Name == "Content" {
				if len(st.Fields.List) != 1 {
					t.Error("Book.Content should only have 1 field named Text")
				}
				for _, e := range st.Fields.List {
					if len(e.Names) != 1 {
						t.Error("Book.Content should only have 1 field named Text")
					} else if e.Names[0].Name != "Text" {
						t.Error("Book.Content should only have 1 field named Text")
					}
					if e.Type.(*ast.Ident).Name != "string" {
						t.Error("Book.Content should only have 1 string field named Text")
					}
				}
			}
		}
		return nil
	}
	cfg := packages.Config{}
	cfg.Mode = packages.NeedSyntax | packages.NeedFiles | packages.NeedTypes | packages.NeedTypesInfo
	err := v.Walk(&cfg, "github.com/hauntedness/packagex")
	if err != nil {
		t.Error()
	}
}
