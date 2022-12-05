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

func (v Visitor) Walk(cfg *packages.Config, pattern ...string) error {
	if cfg == nil {
		cfg = &packages.Config{}
		cfg.Mode = packages.NeedSyntax | packages.NeedFiles | packages.NeedTypes | packages.NeedTypesInfo
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
