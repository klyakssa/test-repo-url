package linter

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "linter",
	Doc:      "checker for panic() calls and log.Fatal/os.Exit outside main.main",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	isInMainMain := func(pos token.Pos) bool {
		for _, file := range pass.Files {
			var found bool
			ast.Inspect(file, func(n ast.Node) bool {
				if fd, ok := n.(*ast.FuncDecl); ok {
					if fd.Body != nil && pos >= fd.Pos() && pos <= fd.End() {
						if fd.Name.Name == "main" && pass.Pkg.Name() == "main" {
							found = true
							return false
						}
					}
				}
				return true
			})
			if found {
				return true
			}
		}
		return false
	}

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		pos := call.Pos()

		if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "panic" {
			pass.Reportf(pos, "avoid using panic() in production code")
			return
		}

		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if xident, ok := sel.X.(*ast.Ident); ok {
				if xident.Name == "log" && (sel.Sel.Name == "Fatal" || sel.Sel.Name == "Fatalf") {
					if !isInMainMain(pos) {
						pass.Reportf(pos, "log.%s should not be used outside main.main", sel.Sel.Name)
					}
					return
				}

				if xident.Name == "os" && sel.Sel.Name == "Exit" {
					if !isInMainMain(pos) {
						pass.Reportf(pos, "os.Exit should not be used outside main.main")
					}
					return
				}
			}
		}
	})

	return nil, nil
}
