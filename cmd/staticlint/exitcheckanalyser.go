package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check that main package don't use os.Exit",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name == "main" {
			ast.Inspect(file, func(node ast.Node) bool {
				switch x := node.(type) {
				case *ast.CallExpr:
					if s, ok := x.Fun.(*ast.SelectorExpr); ok {
						if i, ok := s.X.(*ast.Ident); ok {
							if i.Name == "os" && s.Sel.Name == "Exit" {
								pass.Reportf(x.Pos(), "os.Exit is prohibited in main")
							}
						}
					}
				}
				return true
			})
		}
	}
	return nil, nil
}

//func run(pass *analysis.Pass) (interface{}, error) {
//	for _, file := range pass.Files {
//		if file.Name.Name == "main" {
//			for _, decl := range file.Decls {
//				if fd, ok := decl.(*ast.FuncDecl); ok {
//					if fd.Name.Obj.Name == "main" {
//						ast.Inspect(fd, func(node ast.Node) bool {
//							switch x := node.(type) {
//							case *ast.CallExpr:
//								if s, ok := x.Fun.(*ast.SelectorExpr); ok {
//									if i, ok := s.X.(*ast.Ident); ok {
//										if i.Name == "os" && s.Sel.Name == "Exit" {
//											pass.Reportf(x.Pos(), "os.Exit is prohibited in main")
//										}
//									}
//								}
//							}
//							return true
//						})
//					}
//				}
//			}
//		}
//	}
//	return nil, nil
//}
