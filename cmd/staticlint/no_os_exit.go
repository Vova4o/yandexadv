package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var noOsExitAnalyzer = &analysis.Analyzer{
	Name: "noOsExit",
	Doc:  "check for os.Exit calls in main function of main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if pass.Pkg.Name() != "main" {
			continue
		}

		// Игнорирование сгенерированного кода
		if isGenerated(file) {
			continue
		}

		for _, decl := range file.Decls {
			if fn, isFn := decl.(*ast.FuncDecl); isFn && fn.Name.Name == "main" {
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					if call, isCall := n.(*ast.CallExpr); isCall {
						if sel, isSel := call.Fun.(*ast.SelectorExpr); isSel {
							if pkg, isPkg := sel.X.(*ast.Ident); isPkg && pkg.Name == "os" && sel.Sel.Name == "Exit" {
								pass.Reportf(call.Pos(), "os.Exit call is not allowed in main function")
							}
						}
					}
					return true
				})
			}
		}
	}
	return nil, nil
}

// isGenerated проверяет, является ли файл сгенерированным
func isGenerated(file *ast.File) bool {
	for _, comment := range file.Comments {
		for _, c := range comment.List {
			if c.Text == "// Code generated by 'go test'. DO NOT EDIT." {
				return true
			}
		}
	}
	return false
}
