// Analyzer set of analyzers for static analysis
// Contains:
// printf.Analyzer - check consistency of Printf format strings and arguments.
// shadow.Analyzer - check for possible unintended shadowing of variables.
// structtag.Analyzer - check that struct field tags conform to reflect.StructTag.Get.
// structtag.Analyzer - check that struct field tags conform to reflect.StructTag.Get.
// assign.Analyzer - check for useless assignments.
// atomic.Analyzer - check for common mistakes using the sync/atomic package.
// bools.Analyzer - check for common mistakes involving boolean operators.
// composite.Analyzer - check for unkeyed composite literals.
// copylock.Analyzer - check for locks erroneously passed by value.
// errorsas.Analyzer - report passing non-pointer or non-error values to errors.As.
// httpresponse.Analyzer - check for mistakes using HTTP responses.
// loopclosure.Analyzer - check references to loop variables from within nested functions.
// nilfunc.Analyzer - check for useless comparisons between functions and nil.
// shift.Analyzer - check for shifts that equal or exceed the width of the integer.
// sortslice.Analyzer - check the argument type of sort.Slice.
// stringintconv.Analyzer - check for string(int) conversions.
// tests.Analyzer - check for common mistaken usages of tests and examples.
// unmarshal.Analyzer - report passing non-pointer or non-interface values to unmarshal.
// ExitCheckAnalyzer - checks for a call os.Exit() in main method.
package main

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"honnef.co/go/tools/staticcheck"
	"strings"
)

// ExitCheckAnalyzer checks for a call os.Exit() in main method.
var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check for wrong exit in main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name == "main" {
			ast.Inspect(file, func(node ast.Node) bool {
				switch x := node.(type) {
				case *ast.FuncDecl:
					if x.Name.Name == "main" {
						for _, stmt := range x.Body.List {
							if currSt, ok := stmt.(*ast.ExprStmt); ok {
								if expr, ok := currSt.X.(*ast.CallExpr); ok {
									if f, ok := expr.Fun.(*ast.SelectorExpr); ok {
										funcName := f.X.(*ast.Ident)
										if funcName.Name == "os" && f.Sel.Name == "Exit" {
											pass.Reportf(f.Pos(), "Bad call exit func.")
										}
									}
								}
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

func main() {
	myChecks := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		nilfunc.Analyzer,
		shift.Analyzer,
		sortslice.Analyzer,
		stringintconv.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		ExitCheckAnalyzer,
	}
	for _, v := range staticcheck.Analyzers {
		if strings.Contains(v.Analyzer.Name, "SA") || v.Analyzer.Name == "S1008" {
			myChecks = append(myChecks, v.Analyzer)
		}
	}

	multichecker.Main(
		myChecks...,
	)
}
