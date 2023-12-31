package osexitmain

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer is custom analyzer to use in multichecker. Analyzer checking for os.Exit calls in main function of main package
var Analyzer = &analysis.Analyzer{
	Name: "osexitmain",
	Doc:  "check for os.Exit calls in main function of main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if !validFile(pass.Fset, file) {
			continue
		}

		inMain := false

		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.FuncDecl:
				inMain = x.Name.Name == "main"
			case *ast.CallExpr:
				if !inMain {
					break
				}

				fun, ok := x.Fun.(*ast.SelectorExpr)
				if !ok || fun.Sel.Name != "Exit" {
					break
				}

				pkgIdent, ok := fun.X.(*ast.Ident)
				if !ok || pkgIdent.Name != "os" {
					break
				}

				pass.Reportf(x.Pos(), "found os.Exit in main function")
			}

			return true
		})
	}

	return nil, nil
}

func validFile(fileSet *token.FileSet, file *ast.File) bool {
	if file.Name.Name != "main" {
		return false
	}

	if strings.Contains(fileSet.Position(file.Pos()).Filename, "go-build") {
		return false
	}

	return true
}
