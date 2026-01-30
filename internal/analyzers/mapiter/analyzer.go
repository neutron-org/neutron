package main

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "mapiter",
	Doc:  "flags iteration over maps (non-deterministic; sort keys first)",
	Run: func(pass *analysis.Pass) (any, error) {
		for _, f := range pass.Files {
			filename := pass.Fset.Position(f.Pos()).Filename
			if strings.HasSuffix(filename, "_test.go") {
				continue // skip this file entirely
			}
			ast.Inspect(f, func(n ast.Node) bool {
				r, ok := n.(*ast.RangeStmt)
				if !ok {
					return true
				}
				t := pass.TypesInfo.TypeOf(r.X)
				if t != nil {
					if _, ok := t.Underlying().(*types.Map); ok {
						pass.Reportf(r.For, "iteration over map is non-deterministic; collect and sort keys first")
					}
				}
				return true
			})
		}
		return nil, nil
	},
}
