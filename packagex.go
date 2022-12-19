package packagex

import (
	"errors"
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/packages"
)

type FileContext struct {
	File    *ast.File
	Package *packages.Package
}

func (ctx FileContext) FileName() string {
	return ctx.Package.Fset.File(ctx.File.Pos()).Name()
}

type Context struct {
	FileContext
	GenDecl *ast.GenDecl
}

var ErrNotTypeSpec = errors.New("not a type spec")

func (ctx Context) TypeName() string {
	ts, ok := ctx.GenDecl.Specs[0].(*ast.TypeSpec)
	if ok {
		return ts.Name.Name
	}
	return ""
}

func (ctx Context) TypeParams() (fieldList *ast.FieldList) {
	fs, ok := ctx.GenDecl.Specs[0].(*ast.TypeSpec)
	if ok {
		return fs.TypeParams
	}
	return nil
}

type Visitor struct {
	ValueVisitor        func(*ast.ValueSpec, Context) error
	ImportVisitor       func(*ast.ImportSpec, Context) error
	StructVisitor       func(*ast.StructType, Context) error
	IdentVisitor        func(*ast.Ident, Context) error
	CompositeLitVisitor func(*ast.CompositeLit, Context) error
	fileFilter          func(FileContext) bool
}

func (v *Visitor) WithFileFilter(f func(FileContext) bool) *Visitor {
	v.fileFilter = f
	return v
}

// Walk walks the packages matched with the same rule as [packages]: https://pkg.go.dev/golang.org/x/tools/go/packages,
// calling inside visitor function for each top declaration in each file of the packages
//
//	# Walk will return immediately once any visitor function have error
func (v *Visitor) Walk(cfg *packages.Config, pattern ...string) error {
	if cfg == nil {
		cfg = &packages.Config{}
		cfg.Mode = packages.NeedSyntax | packages.NeedFiles | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedName
	}
	packageList, err := packages.Load(cfg, pattern...)
	if err != nil {
		return err
	}
	for _, pkg := range packageList {
		for _, file := range pkg.Syntax {
			if v.fileFilter != nil && !v.fileFilter(FileContext{File: file, Package: pkg}) {
				continue
			}
			for _, decl := range file.Decls {
				switch typ := decl.(type) {
				case *ast.GenDecl:
					for _, sp := range typ.Specs {
						if spec, ok := sp.(*ast.ValueSpec); ok && v.ValueVisitor != nil {
							err = v.ValueVisitor(spec, Context{FileContext{file, pkg}, typ})
							if err != nil {
								return fmt.Errorf("error call Visitor.ValueVisitor: %w", err)
							}
						} else if spec, ok := sp.(*ast.ImportSpec); ok && v.ImportVisitor != nil {
							err = v.ImportVisitor(spec, Context{FileContext{file, pkg}, typ})
							if err != nil {
								return fmt.Errorf("error call Visitor.ImportVisitor: %w", err)
							}
						} else if spec, ok := sp.(*ast.TypeSpec); ok {
							if st, ok := spec.Type.(*ast.StructType); ok && v.StructVisitor != nil {
								err := v.StructVisitor(st, Context{FileContext{file, pkg}, typ})
								if err != nil {
									return fmt.Errorf("error call Visitor.StructVisitor: %w", err)
								}
							} else if ident, ok := spec.Type.(*ast.Ident); ok && v.IdentVisitor != nil {
								err := v.IdentVisitor(ident, Context{FileContext{file, pkg}, typ})
								if err != nil {
									return fmt.Errorf("error call Visitor.StructVisitor: %w", err)
								}
							} else if cl, ok := spec.Type.(*ast.CompositeLit); ok && v.CompositeLitVisitor != nil {
								err := v.CompositeLitVisitor(cl, Context{FileContext{file, pkg}, typ})
								if err != nil {
									return fmt.Errorf("error call Visitor.StructVisitor: %w", err)
								}
							}
						}
					}
				case *ast.FuncDecl:
				default:
					return fmt.Errorf("not a declare: %v", typ)
				}
			}
		}
	}
	return nil
}

// FieldName return the name of the struct field
func FieldType(expr ast.Expr) string {
	if starExpr, ok := expr.(*ast.StarExpr); ok {
		if selector, ok := starExpr.X.(*ast.SelectorExpr); ok {
			// type = *big.Int
			var pkg = selector.X.(*ast.Ident).Name
			var typ = selector.Sel.Name
			return "*" + pkg + "." + typ
		} else if ident, ok := starExpr.X.(*ast.Ident); ok {
			// type = *int
			return "*" + ident.Name
		} else if star, ok := starExpr.X.(*ast.StarExpr); ok {
			// type = **big.Int
			return "*" + FieldType(star)
		}
	} else if selector, ok := expr.(*ast.SelectorExpr); ok {
		// type = big.Int
		var pkg = selector.X.(*ast.Ident).Name
		var typ = selector.Sel.Name
		return pkg + "." + typ
	} else if ident, ok := expr.(*ast.Ident); ok {
		// type = int
		return ident.Name
	} else if array, ok := expr.(*ast.ArrayType); ok {
		// type = []*big.Int
		return "[]" + FieldType(array.Elt)
	} else if mp, ok := expr.(*ast.MapType); ok {
		// type = map[int]int
		return "map" + "[" + FieldType(mp.Key) + "]" + FieldType(mp.Value)
	} else if ch, ok := expr.(*ast.ChanType); ok {
		if ch.Dir == ast.RECV {
			// type = chan<- int
			return "chan<-" + " " + FieldType(ch.Value)
		}
		if ch.Dir == ast.SEND {
			// type = <-chan int
			return "<-chan" + " " + FieldType(ch.Value)
		}
		// type = chan int
		return "chan" + " " + FieldType(ch.Value)
	}
	return ""
}

var ErrTagNotFound = errors.New("tag not found")

// Tagvalue extract tag value
func TagValue(tag string, key string, subkey string) (string, error) {
	//`json:"name,omitempty" gorm:"column:book;type:varchar(255);"`
	kvs := strings.Split(tag, " ")
	for i := range kvs {
		kv := strings.Trim(kvs[i], `"`)
		before, after, found := strings.Cut(kv, ":")
		if !found || before != key {
			continue
		}
		if subkey == "" {
			return after, nil
		} else {
			return subTagValue(after, subkey)
		}
	}
	return "", ErrTagNotFound
}

func subTagValue(tag string, key string) (string, error) {
	kvs := strings.Split(tag, ";")
	for j := range kvs {
		kv := strings.Trim(kvs[j], " ")
		before, after, found := strings.Cut(kv, ":")
		if !found || before != key {
			continue
		}
		return after, nil
	}
	return "", ErrTagNotFound
}
