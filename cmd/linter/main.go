package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/singlechecker"
	"golang.org/x/tools/go/ast/inspector"
)

// An ForbidExitAnalyzer describes an analysis function and its options.
var ForbidExitAnalyzer = &analysis.Analyzer{
	Name:     "forbidexit",
	Doc:      "reports calls to panic, log.Fatal, and os.Exit outside of main.main",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func main() {
	singlechecker.Main(ForbidExitAnalyzer)
}

func run(pass *analysis.Pass) (any, error) {
	isMainPkg := pass.Pkg.Name() == "main"
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Ищем все вызовы функций (CallExpr)
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	inspect.WithStack(nodeFilter, func(n ast.Node, push bool, stack []ast.Node) bool {
		if !push {
			return true // Обрабатываем только при входе в узел
		}

		call := n.(*ast.CallExpr)

		// Проверка для panic
		if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "panic" {
			pass.Reportf(call.Pos(), "use of built-in function panic is forbidden")
			return true
		}

		// Проверка для log.Fatal и os.Exit
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			pkgIdent, isPkg := sel.X.(*ast.Ident)
			if !isPkg {
				return true
			}

			pkgObj, ok := pass.TypesInfo.Uses[pkgIdent].(*types.PkgName)
			if !ok {
				return true
			}

			funcName := sel.Sel.Name
			pkgName := pkgObj.Imported().Name()

			// Проверяем, находится ли вызов внутри main.main
			if isInsideMainMain(isMainPkg, stack) {
				return true
			}

			if (pkgName == "log" && funcName == "Fatal") ||
				(pkgName == "os" && funcName == "Exit") {
				pass.Reportf(call.Pos(), "call to %s.%s is forbidden outside main.main", pkgName, funcName)
			}
		}

		return true
	})

	return nil, nil
}

// isInsideMainMain checks if the current call stack is inside the main function of the main package.
func isInsideMainMain(isMainPkg bool, stack []ast.Node) bool {
	if !isMainPkg {
		return false
	}

	for _, node := range stack {
		if funcDecl, ok := node.(*ast.FuncDecl); ok {
			if funcDecl.Name.Name == "main" {
				return true
			}
		}
	}
	return false
}
