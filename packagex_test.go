package packagex

import (
	"go/ast"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestVisitor_Walk(t *testing.T) {
	v := Visitor{}
	v.StructVisitor = func(st *ast.StructType, _ Context) error {
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

func TestTagValue(t *testing.T) {
	var visitor Visitor
	visitor.StructVisitor = func(st *ast.StructType, context Context) error {
		if len(context.GenDecl.Specs) != 1 {
			return nil
		}
		structName := context.TypeName()
		if structName != "BookWithTag" {
			return nil
		}
		var fields = st.Fields.List
		if len(fields) != 2 {
			t.Error("expect length 2, got ", len(fields))
		}
		tag0 := fields[0].Tag.Value
		value0, err := TagValue(tag0, "gorm", "type")
		if err != nil {
			t.Error(err)
		}
		if value0 != "varchar(255)" {
			t.Error("expect varchar(255)", "got ", value0)
		}
		tag1 := fields[1].Tag.Value
		value1, err := TagValue(tag1, "gorm", "type")
		if err != nil {
			t.Error(err)
		}
		if value1 != "varchar(30)" {
			t.Error("expect varchar(30)", "got ", value0)
		}
		return nil
	}
	err := visitor.Walk(nil, "github.com/hauntedness/packagex", `.+data.go`)
	if err != nil {
		t.Error(err)
	}
}

func TestContext_FileName(t *testing.T) {
	var visitor Visitor
	visitor.ImportVisitor = func(_ *ast.ImportSpec, ctx Context) error {
		fileName := ctx.FileName()
		if !strings.Contains(fileName, "packagex.go") {
			t.Error("expect file", "got ", fileName)
		}
		return nil
	}
	err := visitor.WithFileFilter(func(ctx FileContext) bool {
		fileName := ctx.FileName()
		if strings.Contains(fileName, "packagex.go") {
			return true
		}
		return false
	}).Walk(nil, "github.com/hauntedness/packagex")
	if err != nil {
		t.Error(err)
	}
}

func TestTypeParam(t *testing.T) {
	var visitor Visitor
	visitor.StructVisitor = func(_ *ast.StructType, ctx Context) error {
		if strings.Contains(ctx.TypeName(), "TypeParam") {
			fieldList := ctx.TypeParams()
			if fieldList.NumFields() != 1 {
				t.Error("error get type params")
			}
			return nil
		}
		return nil
	}
	err := visitor.WithFileFilter(func(ctx FileContext) bool {
		fileName := ctx.FileName()
		if strings.Contains(fileName, "data.go") {
			return true
		}
		return false
	}).Walk(nil, "github.com/hauntedness/packagex")
	if err != nil {
		t.Error(err)
	}
}
