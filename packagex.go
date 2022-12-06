package packagex

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/packages"
)

type Visitor struct {
	ValueVisitor        func(*ast.ValueSpec, *ast.GenDecl, *ast.File, *packages.Package) error
	ImportVisitor       func(*ast.ImportSpec, *ast.GenDecl, *ast.File, *packages.Package) error
	StructVisitor       func(*ast.StructType, *ast.GenDecl, *ast.File, *packages.Package) error
	IdentVisitor        func(*ast.Ident, *ast.GenDecl, *ast.File, *packages.Package) error
	CompositeLitVisitor func(*ast.CompositeLit, *ast.GenDecl, *ast.File, *packages.Package) error
}

// Walk walks the packages matched with the same rule as [packages]: https://pkg.go.dev/golang.org/x/tools/go/packages,
// calling inside visitor function for each top declaration in each file of the packages
//
//	# Walk will return immediately once any visitor function have error
func (v Visitor) Walk(cfg *packages.Config, pattern ...string) error {
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
			for _, decl := range file.Decls {
				switch typ := decl.(type) {
				case *ast.GenDecl:
					for _, sp := range typ.Specs {
						if spec, ok := sp.(*ast.ValueSpec); ok && v.ValueVisitor != nil {
							err = v.ValueVisitor(spec, typ, file, pkg)
							if err != nil {
								return fmt.Errorf("error call Visitor.ValueVisitor: %w", err)
							}
						} else if spec, ok := sp.(*ast.ImportSpec); ok && v.ImportVisitor != nil {
							err = v.ImportVisitor(spec, typ, file, pkg)
							if err != nil {
								return fmt.Errorf("error call Visitor.ImportVisitor: %w", err)
							}
						} else if spec, ok := sp.(*ast.TypeSpec); ok {
							if st, ok := spec.Type.(*ast.StructType); ok && v.StructVisitor != nil {
								err := v.StructVisitor(st, typ, file, pkg)
								if err != nil {
									return fmt.Errorf("error call Visitor.StructVisitor: %w", err)
								}
							} else if ident, ok := spec.Type.(*ast.Ident); ok && v.IdentVisitor != nil {
								err := v.IdentVisitor(ident, typ, file, pkg)
								if err != nil {
									return fmt.Errorf("error call Visitor.StructVisitor: %w", err)
								}
							} else if cl, ok := spec.Type.(*ast.CompositeLit); ok && v.CompositeLitVisitor != nil {
								err := v.CompositeLitVisitor(cl, typ, file, pkg)
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
func FieldName(expr ast.Expr) string {
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
			return "*" + FieldName(star)
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
		return "[]" + FieldName(array.Elt)
	}
	return ""
}
